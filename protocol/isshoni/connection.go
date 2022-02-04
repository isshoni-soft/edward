package isshoni

import (
	"fmt"
	"github.com/isshoni-soft/edward/packet"
)

type ConnectPacket struct {
	Name    string
	Version string
}

type AcceptConnectionPacket struct{}

type DisconnectPacket struct{}

func Connect(protocol *Protocol) func(c packet.Channel, dataI interface{}) {
	return func(c packet.Channel, dataI interface{}) {
		data := dataI.(*ConnectPacket)
		channelData := protocol.ChannelMeta().Fetch(c)

		fmt.Println("Received handshake on channel  " + c.UUID().String())
		fmt.Println("Handshake data:", dataI)

		if data.Name != Name || data.Version != Version {
			fmt.Println("Incompatible Protocol Versions:")
			fmt.Println("Client: " + data.Name + " | v" + data.Version)
			fmt.Println("Ours: " + Name + " | v" + Version)
			c.SendPacket(DisconnectPacket{})

			c.Close()
			return
		}

		channelData.Handshake = true

		fmt.Println("Accepting connection...")
		c.SendPacket(AcceptConnectionPacket{})
	}
}

func AcceptConnection(protocol *Protocol) func(c packet.Channel, dataI interface{}) {
	return func(c packet.Channel, dataI interface{}) {
		channelData := protocol.ChannelMeta().Fetch(c)

		channelData.Handshake = true

		fmt.Println("Successfully connected.")
	}
}

func Disconnect() func(c packet.Channel, dataI interface{}) {
	return func(c packet.Channel, dataI interface{}) {
		fmt.Println("Received disconnect packet.")
		c.Close()
	}
}
