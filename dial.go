package network

import (
	"github.com/isshoni-soft/edward/packet"
	"github.com/isshoni-soft/edward/protocol"
	"net"
)

func Dial(address string, port string, encoder packet.Encoder, protocol protocol.Manager) *packet.Channel {
	var connection net.Conn

	protocol.RegisterPackets(encoder.PacketRegistry())

	if c, err := net.Dial("tcp", address+":"+port); err != nil {
		panic(err)
	} else {
		connection = c
	}

	result := packet.NewChannel(connection, encoder)
	result.Start()

	return result
}
