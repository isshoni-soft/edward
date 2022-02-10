package packet

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/isshoni-soft/edward"
	"reflect"
)

type Encoder interface {
	EncodePacket(data interface{}) (string, error)
	DecodePacket(str string) (*DecodedPacket, error)
	PacketRegistry() Registry
}

type packet struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
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
		registry: NewSimpleRegistry(),
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
		return "", network.UnrecognizedPacket{
			Type: reflect.TypeOf(data).String(),
		}
	}

	packet := packet{
		Name: s.registry.GetPacketNameByType(reflect.TypeOf(data)),
		Data: data,
	}

	if d, err := json.Marshal(packet); err == nil {
		fmt.Println("Final packet: " + string(d))

		// TODO: Encrypt here

		return base64.StdEncoding.EncodeToString(d), nil
	} else {
		return "", err
	}
}

func (s *SimpleEncoder) DecodePacket(str string) (*DecodedPacket, error) {
	var decodedStr string
	var raw []byte

	fmt.Print("Decoding packet: " + str)

	if d, err := base64.StdEncoding.DecodeString(str); err == nil {
		decodedStr = string(d)
		raw = d
	} else {
		return nil, network.MalformedProtocol{
			Reason: "failed to decode base64",
		}
	}

	fmt.Println("Decoded from base64: " + decodedStr)

	// TODO: Decrypt here

	var packet *packet

	fmt.Println("Unmarshalling packet...")
	if err := json.Unmarshal(raw, &packet); err != nil {
		return nil, err
	}

	var packetType reflect.Type

	if !s.registry.IsPacketName(packet.Name) {
		return nil, network.UnrecognizedPacket{
			Type: packet.Name,
		}
	} else {
		packetType = s.registry.GetPacketTypeByName(packet.Name)
	}

	// this might be overkill, but I don't trust the json library
	// want to make sure we can safely cast the received interface to the packet's type
	packetValue := reflect.New(packetType).Interface()

	remarshaled, _ := json.Marshal(packet.Data)

	fmt.Println("RE: " + string(remarshaled))

	if err := json.Unmarshal(remarshaled, &packetValue); err != nil {
		fmt.Println("ERROR: ", err)
	}

	result := &DecodedPacket{
		Name: packet.Name,
		Type: packetType,
		Data: packetValue,
	}

	return result, nil
}

func (s *SimpleEncoder) PacketRegistry() Registry {
	return s.registry
}
