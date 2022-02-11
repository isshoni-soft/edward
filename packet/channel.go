package packet

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	errors "github.com/isshoni-soft/edward/error"
	"io"
	"net"
	"reflect"
)

// Packet System
// TODO: handles encryption

type Listener func(channel Channel, packet interface{})

type Channel interface {
	Start()
	Close()
	SetCloseCallback(callback func(c Channel))
	RegisterPacketListener(sample interface{}, listener Listener)
	SendPacket(packet interface{}) (chan bool, error)
	SendRawMessage(str string) chan bool
	UUID() uuid.UUID
	Running() bool
}

type outboundPacket struct {
	Sent    chan bool
	Message string
}

type SimpleChannel struct {
	Encoder Encoder
	Uuid    uuid.UUID
	OnClose func(c Channel)

	connection     net.Conn
	running        bool
	outShutdown    chan bool
	inShutdown     chan bool
	readerShutdown chan bool
	outbound       chan outboundPacket
	inbound        chan string
	listeners      map[reflect.Type][]Listener
}

func NewChannel(connection net.Conn, encoder Encoder) Channel {
	return &SimpleChannel{
		Encoder:        encoder,
		Uuid:           uuid.New(),
		OnClose:        func(c Channel) {},
		connection:     connection,
		running:        false,
		outShutdown:    make(chan bool),
		inShutdown:     make(chan bool),
		readerShutdown: make(chan bool, 2), // buffer this so reader doesn't block when submitting close on error
		outbound:       make(chan outboundPacket, 20),
		inbound:        make(chan string, 20),
		listeners:      make(map[reflect.Type][]Listener),
	}
}

func (c *SimpleChannel) Start() {
	c.running = true

	c.run(c.readerFunc)
	c.run(c.inboundFunc)
	c.run(c.outboundFunc)
}

func (c *SimpleChannel) Close() {
	if !c.running {
		return
	}

	fmt.Println("Closing channel: " + c.UUID().String())

	c.running = false

	fmt.Println("Closing connection...")
	c.connection.Close()

	fmt.Println("Waiting for reader thread to shutdown")
	<-c.readerShutdown

	close(c.inbound)
	fmt.Println("Waiting for inbound thread to shutdown")
	<-c.inShutdown

	close(c.outbound)
	fmt.Println("Waiting for outbound thread to shutdown")
	<-c.outShutdown

	c.OnClose(c)
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

func (c *SimpleChannel) SendPacket(packet interface{}) (chan bool, error) {
	if !c.running {
		return nil, errors.ClosedChannel{}
	}

	fmt.Println("Sending packet", packet, "on", c.Uuid)

	str, err := c.Encoder.EncodePacket(packet)

	fmt.Println("Encoded packet: " + str)

	if err != nil {
		return nil, err
	}

	return c.SendRawMessage(str), nil
}

func (c *SimpleChannel) SendRawMessage(str string) chan bool {
	p := outboundPacket{
		Sent:    make(chan bool, 2),
		Message: str,
	}

	c.outbound <- p

	return p.Sent
}

func (c SimpleChannel) UUID() uuid.UUID {
	return c.Uuid
}

func (c SimpleChannel) Running() bool {
	return c.running
}

func (c *SimpleChannel) outboundFunc() {
	for p := range c.outbound {
		fmt.Fprintln(c.connection, p.Message)
		p.Sent <- true
	}

	c.outShutdown <- true
}

func (c *SimpleChannel) readerFunc() {
	reader := bufio.NewReader(c.connection)

	for {
		var message string
		var err error

		if message, err = reader.ReadString('\n'); err != nil {
			if err == io.EOF || !c.running {
				c.readerShutdown <- true

				if c.running {
					c.Close()
				}

				return
			}

			fmt.Println("[error]: failed to read channel: " + c.UUID().String())
			fmt.Println(err)

			continue
		}

		c.inbound <- message
	}
}

func (c *SimpleChannel) inboundFunc() {
	for message := range c.inbound {
		fmt.Println("Processing inbound: " + message)

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
				go listener(c, decoded.Data)
			}
		}

		fmt.Println("Finished processing packet.")
	}

	fmt.Println("Submitting shutdown...")
	c.inShutdown <- true
}

func (c SimpleChannel) run(f func()) {
	go func() {
		f()
	}()
}
