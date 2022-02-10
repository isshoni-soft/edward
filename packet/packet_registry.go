package packet

import "reflect"

type StrictPacket interface {
	Valid() bool
}

type Registry interface {
	RegisterPacket(name string, sample interface{})
	GetPacketTypeByName(name string) reflect.Type
	GetPacketNameByType(t reflect.Type) string
	IsPacket(data interface{}) bool
	IsPacketName(name string) bool
}

type SimpleRegistry struct {
	registry        map[string]reflect.Type
	reverseRegistry map[reflect.Type]string
}

func NewSimpleRegistry() *SimpleRegistry {
	return &SimpleRegistry{
		registry:        make(map[string]reflect.Type),
		reverseRegistry: make(map[reflect.Type]string),
	}
}

func (s *SimpleRegistry) RegisterPacket(name string, sample interface{}) {
	s.registry[name] = reflect.TypeOf(sample)
	s.reverseRegistry[reflect.TypeOf(sample)] = name
}

func (s *SimpleRegistry) GetPacketTypeByName(name string) reflect.Type {
	return s.registry[name]
}

func (s *SimpleRegistry) GetPacketNameByType(t reflect.Type) string {
	return s.reverseRegistry[t]
}

func (s *SimpleRegistry) IsPacket(data interface{}) (ok bool) {
	_, ok = s.reverseRegistry[reflect.TypeOf(data)]
	return
}

func (s *SimpleRegistry) IsPacketName(name string) (ok bool) {
	_, ok = s.registry[name]
	return
}
