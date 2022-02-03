package protocol

import "github.com/isshoni-soft/edward/packet"

const (
	Version = "0.1.0"
	Name    = "Isshoni"
)

type Manager interface {
	RegisterPackets(registry packet.Registry)
	Version() string
	Name() string
}

type IsshoniProtocol struct{}

func (i IsshoniProtocol) RegisterPackets(registry packet.Registry) {
	registry.RegisterPacket("handshake", HandshakePacket{})
}

func (i IsshoniProtocol) Version() string {
	return Version
}

func (i IsshoniProtocol) Name() string {
	return Name
}
