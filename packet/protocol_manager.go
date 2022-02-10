package packet

type Manager interface {
	RegisterPackets(registry Registry)
	RegisterListeners(channel Channel)
	ClientPostStart(channel Channel)
}
