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

type InvalidStrictPacket struct {
	Data interface{}
}

func (i InvalidStrictPacket) Error() string {
	return "strict packet failed validation"
}

type MalformedPacket struct {
	Data interface{}
}

func (m MalformedPacket) Error() string {
	return "packet data could not be unmarshaled into packet struct"
}
