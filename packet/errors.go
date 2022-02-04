package packet

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
