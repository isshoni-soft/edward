package packet

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	network "github.com/isshoni-soft/edward/error"
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
	SendPacket(packet interface{}) error
	SendRawMessage(str string)
	UUID() uuid.UUID
	Running() bool
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
	outbound       chan string
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
		outbound:       make(chan string, 20),
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
	fmt.Println("Closing channel: " + c.UUID().String())

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

	c.running = false

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

func (c *SimpleChannel) SendPacket(packet interface{}) error {
	if !c.running {
		return network.ClosedChannel{}
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
	for message := range c.outbound {
		fmt.Fprintln(c.connection, message)
	}

	c.outShutdown <- true
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
				c.readerShutdown <- true
				c.Close()
				return
			}

			continue
		}

		c.inbound <- message
	}
}

func (c *SimpleChannel) inboundFunc() {
	for message := range c.inbound {
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
	}

	c.inShutdown <- true
}

func (c SimpleChannel) run(f func()) {
	go func() {
		f()
	}()
}
