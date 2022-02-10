package error

type UnrecognizedPacket struct {
	Type string
}

func (u UnrecognizedPacket) Error() string {
	return u.Type + " is not a registered packet"
}

type ClosedChannel struct{}

func (c ClosedChannel) Error() string {
	return "this channel is not running"
}
