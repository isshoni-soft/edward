package packet

type Manager interface {
	RegisterPackets(registry Registry)
	RegisterListeners(channel Channel)
	PostClientStart(channel Channel)
}
