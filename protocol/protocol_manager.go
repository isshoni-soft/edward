package protocol

import (
	"github.com/isshoni-soft/edward/packet"
)

type Manager interface {
	RegisterPackets(registry packet.Registry)
	RegisterListeners(channel packet.Channel)
	ClientPostStart(channel packet.Channel)
}
