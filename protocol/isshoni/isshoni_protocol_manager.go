package isshoni

import (
	"github.com/google/uuid"
	"github.com/isshoni-soft/edward/packet"
)

const (
	Version = "0.1.0"
	Name    = "Isshoni"
)

type Protocol struct {
	channelMeta *ChannelMetadataManager
}

func New() *Protocol {
	return &Protocol{
		channelMeta: NewChannelMetadataManager(),
	}
}

func (i *Protocol) RegisterPackets(registry packet.Registry) {
	registry.RegisterPacket("connect", ConnectPacket{})
	registry.RegisterPacket("accept_connection", AcceptConnectionPacket{})
	registry.RegisterPacket("disconnect", DisconnectPacket{})
	registry.RegisterPacket("message", MessagePacket{})
}

func (i *Protocol) RegisterListeners(channel packet.Channel) {
	i.channelMeta.Register(channel)

	channel.RegisterPacketListener(ConnectPacket{}, Connect(i))
	channel.RegisterPacketListener(DisconnectPacket{}, Disconnect())
	channel.RegisterPacketListener(AcceptConnectionPacket{}, AcceptConnection(i))
}

func (i *Protocol) ClientPostStart(channel packet.Channel) {
	channel.SendPacket(ConnectPacket{
		Name:    Name,
		Version: Version,
	})
}

func (i Protocol) ChannelMeta() *ChannelMetadataManager {
	return i.channelMeta
}

type ChannelMetadata struct {
	channel packet.Channel

	Handshake bool
}

func (c ChannelMetadata) Channel() packet.Channel {
	return c.channel
}

type ChannelMetadataManager struct {
	channels map[uuid.UUID]*ChannelMetadata
}

func NewChannelMetadataManager() *ChannelMetadataManager {
	return &ChannelMetadataManager{
		channels: make(map[uuid.UUID]*ChannelMetadata),
	}
}

func (c *ChannelMetadataManager) Register(channel packet.Channel) {
	c.channels[channel.UUID()] = &ChannelMetadata{
		channel:   channel,
		Handshake: false,
	}
}

func (c *ChannelMetadataManager) Fetch(channel packet.Channel) *ChannelMetadata {
	return c.channels[channel.UUID()]
}
