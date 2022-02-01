package network

import (
	"bufio"
	"fmt"
	"net"
)

// Packet System
// TODO: packet channel is main object for interacting with a connected client
// TODO: packet channel enqueues inbound/outbound messages
// TODO: dedicated threads for inbound/outbound messages
// TODO: handles encryption and key propagation
// TODO: allow packet type registration along with listeners -- better when go 1.18 arrives
// TODO: handles de/serialization into registered packet types
// TODO: simple packets can be automatically serialized, similar to json library serialization

type PacketChannel struct {
	connection net.Conn
	running    bool
	shutdown   chan bool
	outbound   *messageQueue
	inbound    *messageQueue
}

func NewPacketChannel(connection net.Conn) *PacketChannel {
	return &PacketChannel{
		connection: connection,
		running:    false,
		shutdown:   make(chan bool),
		outbound:   newMessageQueue(),
		inbound:    newMessageQueue(),
	}
}

func (p *PacketChannel) Start() chan bool {
	p.run(p.inboundFunc)
	p.run(p.outboundFunc)

	p.running = true

	return p.shutdown
}

func (p *PacketChannel) Close() {
	p.shutdown <- true
}

func (p *PacketChannel) outboundFunc() {
	for {
		select {
		case <-p.shutdown:
			return
		default:
		}

		if p.outbound.Empty() {
			continue
		}

		message, _ := p.outbound.Pop()

		// TODO: Do a little bit of encoding

		fmt.Fprintln(p.connection, message)
	}
}

func (p *PacketChannel) inboundFunc() {
	reader := bufio.NewReader(p.connection)

	for {
		select {
		case <-p.shutdown:
			return
		default:
		}

		var message string

		if m, err := reader.ReadString('\n'); err != nil {
			message = m
		} else {
			fmt.Println("[error]: reader thread errored!")
			continue
		}

		p.inbound.Push(message)
	}
}

func (p PacketChannel) run(f func()) {
	go func() {
		f()
	}()
}
