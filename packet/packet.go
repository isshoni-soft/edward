package packet

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
)

// Packet System
// TODO: channel is main object for interacting with a connected client
// TODO: channel enqueues inbound/outbound messages
// TODO: handles encryption and key propagation
// TODO: allow packet type registration along with listeners -- make better when go 1.18 arrives
// TODO: handles de/serialization into registered packet types
// TODO: simple packets can be automatically serialized, similar to json library serialization

type Listener func(packet *interface{})

type Channel struct {
	Encoder Encoder

	connection net.Conn
	running    bool
	shutdown   chan bool
	outbound   *messageQueue
	listeners  map[reflect.Type][]Listener
}

func NewChannel(connection net.Conn, encoder Encoder) *Channel {
	return &Channel{
		Encoder:    encoder,
		connection: connection,
		running:    false,
		shutdown:   make(chan bool),
		outbound:   newMessageQueue(),
	}
}

func (c *Channel) Start() {
	c.run(c.inboundFunc)
	c.run(c.outboundFunc)

	c.running = true
}

func (c *Channel) Close() {
	c.shutdown <- true
	c.running = false
}

func (c *Channel) RegisterPacketListener(packetType reflect.Type, listener Listener) {
	if s, ok := c.listeners[packetType]; ok {
		c.listeners[packetType] = append(s, listener)
	} else {
		c.listeners[packetType] = []Listener{listener}
	}
}

func (c *Channel) SendPacket(packet interface{}) error {
	str, err := c.Encoder.EncodePacket(packet)

	if err != nil {
		return err
	}

	c.SendRawMessage(str)

	return nil
}

func (c *Channel) SendRawMessage(str string) {
	c.outbound.Push(str)
}

func (c *Channel) outboundFunc() {
	for {
		select {
		case <-c.shutdown:
			return
		default:
		}

		if c.outbound.Empty() {
			continue
		}

		message, _ := c.outbound.Pop()

		fmt.Fprintln(c.connection, message)
	}
}

func (c *Channel) inboundFunc() {
	reader := bufio.NewReader(c.connection)

	for {
		select {
		case <-c.shutdown:
			return
		default:
		}

		var message string

		if m, err := reader.ReadString('\n'); err == nil {
			message = m
		} else {
			fmt.Println("[error]: reader thread errored!")
			fmt.Println(err)
			continue
		}

		// TODO: Decode Base64 to encrypted blob
		// TODO: Un-encrypt blob
		// TODO: Create packet object then fire on packet listeners

		fmt.Print(message)
	}
}

func (c Channel) run(f func()) {
	go func() {
		f()
	}()
}
