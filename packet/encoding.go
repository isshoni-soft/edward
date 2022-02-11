package packet

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	errors "github.com/isshoni-soft/edward/error"
	"reflect"
)

type Encoder interface {
	EncodePacket(data interface{}) (string, error)
	DecodePacket(str string) (*DecodedPacket, error)
	PacketRegistry() Registry
}

type networkPacket struct {
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
		return "", errors.UnrecognizedPacket{
			Type: reflect.TypeOf(data).String(),
		}
	}

	packet := networkPacket{
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
		return nil, errors.ProtocolError{
			Reason: "failed to decode base64",
		}
	}

	fmt.Println("Decoded from base64: " + decodedStr)

	// TODO: Decrypt here

	var packet *networkPacket

	fmt.Println("Unmarshalling packet...")
	if err := json.Unmarshal(raw, &packet); err != nil {
		return nil, err
	}

	var packetType reflect.Type

	if !s.registry.IsPacketName(packet.Name) {
		return nil, errors.UnrecognizedPacket{
			Type: packet.Name,
		}
	} else {
		packetType = s.registry.GetPacketTypeByName(packet.Name)
	}
	
	packetValue := reflect.New(packetType).Interface()

	// we have to actually do this because of the way this data is currently stored in memory
	// rather than manually mapping this different format I opt to remarshal just packet data
	// then unmarshal it into an empty value of the desired type. All of this just so that
	// we can cast the packet data interface.
	remarshaled, _ := json.Marshal(packet.Data)

	if err := json.Unmarshal(remarshaled, &packetValue); err != nil {
		return nil, errors.MalformedPacket{
			Data: packet.Data,
		}
	}

	result := &DecodedPacket{
		Name: packet.Name,
		Type: packetType,
		Data: packetValue,
	}

	if strict, ok := packetValue.(StrictPacket); ok && !strict.Valid() {
		return nil, errors.InvalidStrictPacket{
			Data: packetValue,
		}
	}

	return result, nil
}

func (s *SimpleEncoder) PacketRegistry() Registry {
	return s.registry
}
