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
}

func NewPacketChannel(connection net.Conn) *PacketChannel {
	return &PacketChannel{
		connection: connection,
		running:    false,
		shutdown:   make(chan bool),
		outbound:   newMessageQueue(),
	}
}

func (p *PacketChannel) Start() {
	p.run(p.inboundFunc)
	p.run(p.outboundFunc)

	p.running = true
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

		// TODO: Do a little bit of encryption
		// TODO: Base64 encode the encrypted blob

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

		// TODO: Decode Base64 to encrypted blob
		// TODO: Un-encrypt blob
		// TODO: Create packet object then fire on packet listeners

		fmt.Println(message)
	}
}

func (p PacketChannel) run(f func()) {
	go func() {
		f()
	}()
}
