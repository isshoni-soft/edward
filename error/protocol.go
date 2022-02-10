package error

type NoProtocol struct{}

func (n NoProtocol) Error() string {
	return "network listener requires a protocol"
}

type ProtocolError struct {
	Reason string
}

func (m ProtocolError) Error() string {
	return "malformed protocol: " + m.Reason
}
