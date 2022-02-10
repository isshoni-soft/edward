package network

type UnrecognizedPacket struct {
	Type string
}

func (u UnrecognizedPacket) Error() string {
	return u.Type + " is not a registered packet"
}

type MalformedProtocol struct {
	Reason string
}

func (m MalformedProtocol) Error() string {
	return "malformed protocol: " + m.Reason
}

type ClosedChannel struct{}

func (c ClosedChannel) Error() string {
	return "this channel is not running"
}

type ListenerRunning struct{}

func (l ListenerRunning) Error() string {
	return "this listener is already started"
}

type NoProtocol struct{}

func (n NoProtocol) Error() string {
	return "network listener requires a protocol"
}
