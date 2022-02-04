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
	SetCloseCallback(callback func(c Channel))
	RegisterPacketListener(sample interface{}, listener Listener)
	SendPacket(packet interface{}) error
	SendRawMessage(str string)
	UUID() uuid.UUID
	Running() bool
}

type SimpleChannel struct {
	Encoder Encoder
	Uuid    uuid.UUID
	OnClose func(c Channel)

	connection  net.Conn
	running     bool
	shutdownOut chan bool
	outShutdown chan bool
	shutdownIn  chan bool
	inShutdown  chan bool
	outbound    chan string
	inbound     chan string
	listeners   map[reflect.Type][]Listener
}

func NewChannel(connection net.Conn, encoder Encoder) Channel {
	return &SimpleChannel{
		Encoder:     encoder,
		Uuid:        uuid.New(),
		OnClose:     func(c Channel) {},
		connection:  connection,
		running:     false,
		shutdownOut: make(chan bool),
		outShutdown: make(chan bool),
		shutdownIn:  make(chan bool),
		inShutdown:  make(chan bool),
		outbound:    make(chan string, 10),
		inbound:     make(chan string, 10),
		listeners:   make(map[reflect.Type][]Listener),
	}
}

func (c *SimpleChannel) Start() {
	c.run(c.readerFunc)
	c.run(c.inboundFunc)
	c.run(c.outboundFunc)

	c.running = true
}

func (c *SimpleChannel) Close() {
	fmt.Println("Closing channel: " + c.UUID().String())

	c.shutdownIn <- true
	<-c.inShutdown
	c.shutdownOut <- true
	<-c.outShutdown

	c.running = false

	c.OnClose(c)
	c.connection.Close()
	fmt.Println("Channel closed!")
}

func (c *SimpleChannel) SetCloseCallback(callback func(c Channel)) {
	c.OnClose = callback
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
	if !c.running {
		return ClosedChannel{}
	}

	fmt.Println("Sending packet", packet, "on", c.Uuid)

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

func (c SimpleChannel) Running() bool {
	return c.running
}

func (c *SimpleChannel) outboundFunc() {
	for {
		select {
		case <-c.shutdownOut:
			fmt.Println("Shutting down outbound for: " + c.UUID().String())
			for len(c.outbound) > 0 {
				c.handleOut()
			}
			fmt.Println("Outbound shutdown.")

			c.outShutdown <- true
			return
		default:
		}

		c.handleOut()
	}
}

func (c *SimpleChannel) handleOut() {
	select {
	case message := <-c.outbound:
		fmt.Fprintln(c.connection, message)
	default:
	}
}

func (c *SimpleChannel) readerFunc() {
	reader := bufio.NewReader(c.connection)

	for {
		var message string
		var err error

		if message, err = reader.ReadString('\n'); err != nil {
			fmt.Println("[error]: failed to read channel: " + c.UUID().String())
			fmt.Println(err)

			if err == io.EOF {
				c.Close()
				return
			}

			continue
		}

		c.inbound <- message
	}
}

func (c *SimpleChannel) inboundFunc() {
	for {
		select {
		case <-c.shutdownIn:
			fmt.Println("Shutting down inbound for: " + c.UUID().String())
			for len(c.inbound) > 0 {
				c.handleIn()
			}
			fmt.Println("Inbound shutdown.")

			c.inShutdown <- true
			return
		default:
		}

		c.handleIn()
	}
}

func (c *SimpleChannel) handleIn() {
	select {
	case message := <-c.inbound:
		var err error
		var decoded *DecodedPacket

		if decoded, err = c.Encoder.DecodePacket(message); err != nil {
			fmt.Println("[error]: failed to decode packet on channel: " + c.UUID().String())
			fmt.Println(err)
			return
		}

		fmt.Println("Received packet:", decoded)

		if listeners, ok := c.listeners[decoded.Type]; ok {
			fmt.Println("Sending packet to listeners.")

			for _, listener := range listeners {
				listener(c, decoded.Data)
			}
		}

		fmt.Println("Finished processing packet.")
	default:
	}
}

func (c SimpleChannel) run(f func()) {
	go func() {
		f()
	}()
}
