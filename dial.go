package network

import (
	"fmt"
	"github.com/isshoni-soft/edward/packet"
	"net"
)

type Dialer struct {
	Address  string
	Port     string
	Encoder  packet.Encoder
	Protocol Manager
	PreInit  func(c packet.Channel)
}

func Dial(dialer Dialer) packet.Channel {
	var connection net.Conn

	dialer.Protocol.RegisterPackets(dialer.Encoder.PacketRegistry())

	fmt.Println("Connecting to: " + dialer.Address + ":" + dialer.Port + "...")

	if c, err := net.Dial("tcp", dialer.Address+":"+dialer.Port); err != nil {
		panic(err)
	} else {
		connection = c
	}

	fmt.Println("Establishing connection...")
	result := packet.NewChannel(connection, dialer.Encoder)
	fmt.Println("Pre-initializing channel...")
	dialer.PreInit(result)
	fmt.Println("Initializing channel...")
	result.Start()
	fmt.Println("Protocol post connect.")
	dialer.Protocol.PostClientStart(result)

	fmt.Println("Channel initialized!")

	return result
}
