package network

import "net"

// Packet System
// TODO: packet channel is main object for interacting with a connected client
// TODO: packet channel enqueues inbound/outbound messages
// TODO: dedicated threads for inbound/outbound messages
// TODO: handles encryption and key propagation
// TODO: allow packet type registration along with listeners -- better when go 1.18 arrives
// TODO: handles de/serialization into registered packet types
// TODO: simple packets can be automatically serialized, similar to json library serialization

type Serializer interface {
	Serialize() string
}

type Deserializer interface {
	Deserialize(str string) interface{}
}

type PacketChannel struct {
	Connection net.Conn

	running  bool
	shutdown chan bool
	outbound *messageQueue
	inbound  *messageQueue
}

func (p *PacketChannel) Start() chan bool {
	if p.shutdown == nil {
		p.shutdown = make(chan bool, 1)
	}

	// TODO: Start inbound/outbound threads

	p.running = true

	return p.shutdown
}
