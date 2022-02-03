package packet

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
)

type Encoder interface {
	EncodePacket(data interface{}) (string, error)
	DecodePacket(str string) (*DecodedPacket, error)
	PacketRegistry() Registry
}

type packet struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type DecodedPacket struct {
	Name string
	Type reflect.Type
	Data interface{}
}

type SimpleEncoder struct {
	registry Registry
}

func NewSimpleEncoder() *SimpleEncoder {
	result := &SimpleEncoder{
		registry: &SimpleRegistry{},
	}

	return result
}

func NewSimpleEncoderWithRegistry(registry Registry) *SimpleEncoder {
	return &SimpleEncoder{
		registry: registry,
	}
}

func (s *SimpleEncoder) EncodePacket(data interface{}) (string, error) {
	if !s.registry.IsPacket(data) {
		return "", UnrecognizedPacket{
			Type: reflect.TypeOf(data).String(),
		}
	}

	packet := packet{
		Name: s.registry.GetPacketNameByType(reflect.TypeOf(data)),
	}

	if d, err := json.Marshal(packet); err == nil {

		// TODO: Encrypt here

		return base64.StdEncoding.EncodeToString(d), nil
	} else {
		return "", err
	}
}

func (s *SimpleEncoder) DecodePacket(str string) (*DecodedPacket, error) {
	var decodedStr string

	if d, err := base64.StdEncoding.DecodeString(str); err == nil {
		decodedStr = string(d)
	} else {
		return nil, MalformedProtocol{
			Reason: "failed to decode base64",
		}
	}

	// TODO: Decrypt here

	var packet *packet

	if err := json.Unmarshal([]byte(decodedStr), packet); err != nil {
		return nil, err
	}

	var packetType reflect.Type

	if !s.registry.IsPacketName(packet.Name) {
		return nil, UnrecognizedPacket{
			Type: packet.Name,
		}
	} else {
		packetType = s.registry.GetPacketTypeByName(packet.Name)
	}

	// this might be overkill, but I don't trust the json library
	remarshaled, _ := json.Marshal(packet.Data)

	// want to make sure we can safely cast the received interface to the packet's type
	packetValue := reflect.New(packetType).Elem().Interface()

	if err := json.Unmarshal(remarshaled, &packetValue); err != nil {
		return nil, err
	}

	return &DecodedPacket{
		Name: packet.Name,
		Type: packetType,
		Data: packetValue,
	}, nil
}

func (s *SimpleEncoder) PacketRegistry() Registry {
	return s.registry
}
