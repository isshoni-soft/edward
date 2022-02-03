package protocol

const ProtocolVersion = "0.1.0"

type HandshakePacket struct {
	Name    string
	Version string
}
