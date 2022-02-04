package packet

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"io"
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

type Listener func(channel Channel, packet interface{})

type Channel interface {
	Start()
	Close()
	RegisterPacketListener(sample interface{}, listener Listener)
	SendPacket(packet interface{}) error
	SendRawMessage(str string)
	UUID() uuid.UUID
}

type SimpleChannel struct {
	Encoder Encoder
	Uuid    uuid.UUID

	connection  net.Conn
	running     bool
	shutdownOut chan bool
	shutdownIn  chan bool
	outbound    chan string
	listeners   map[reflect.Type][]Listener
}

func NewChannel(connection net.Conn, encoder Encoder) Channel {
	return &SimpleChannel{
		Encoder:     encoder,
		Uuid:        uuid.New(),
		connection:  connection,
		running:     false,
		shutdownOut: make(chan bool),
		shutdownIn:  make(chan bool),
		outbound:    make(chan string, 10),
		listeners:   make(map[reflect.Type][]Listener),
	}
}

func (c *SimpleChannel) Start() {
	c.run(c.inboundFunc)
	c.run(c.outboundFunc)

	c.running = true
}

func (c *SimpleChannel) Close() {
	c.shutdownOut <- true
	c.shutdownIn <- true
	c.running = false
}

func (c *SimpleChannel) RegisterPacketListener(sample interface{}, listener Listener) {
	packetType := reflect.TypeOf(sample)

	if s, ok := c.listeners[packetType]; ok {
		c.listeners[packetType] = append(s, listener)
	} else {
		c.listeners[packetType] = []Listener{listener}
	}
}

func (c *SimpleChannel) SendPacket(packet interface{}) error {
	fmt.Println("Sending packet", packet)

	str, err := c.Encoder.EncodePacket(packet)

	fmt.Println("Encoded packet: " + str)

	if err != nil {
		return err
	}

	c.SendRawMessage(str)

	return nil
}

func (c *SimpleChannel) SendRawMessage(str string) {
	c.outbound <- str
}

func (c SimpleChannel) UUID() uuid.UUID {
	return c.Uuid
}

func (c *SimpleChannel) outboundFunc() {
	for {
		select {
		case <-c.shutdownOut:
			fmt.Println("Channel: " + c.UUID().String() + " received shutdown, dumping message queue...")
			for len(c.outbound) > 0 {
				c.handleOut()
			}
			fmt.Println("Channel: " + c.UUID().String() + " shutdown")

			return
		default:
		}

		c.handleOut()
	}
}

func (c *SimpleChannel) handleOut() {
	message := <-c.outbound

	fmt.Fprintln(c.connection, message)
}

func (c *SimpleChannel) inboundFunc() {
	reader := bufio.NewReader(c.connection)

	for {
		select {
		case <-c.shutdownIn:
			return
		default:
		}

		var message string
		var err error

		if message, err = reader.ReadString('\n'); err != nil {
			fmt.Println("[error]: failed to read channel: " + c.UUID().String())

			if err == io.EOF {
				c.Close()
			}

			continue
		}

		var decoded *DecodedPacket

		if decoded, err = c.Encoder.DecodePacket(message); err != nil {
			fmt.Println("[error]: failed to decode packet on channel: " + c.UUID().String())
			fmt.Println(err)
			continue
		}

		fmt.Println("Received packet:", decoded)

		if listeners, ok := c.listeners[decoded.Type]; ok {
			fmt.Println("Sending packet to listeners.")

			for _, listener := range listeners {
				listener(c, decoded.Data)
			}
		}

		fmt.Println("Finished processing packet.")
	}
}

func (c SimpleChannel) run(f func()) {
	go func() {
		f()
	}()
}
