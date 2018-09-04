package packets

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	PACKET_TYPE_INPUT uint8 = iota
	PACKET_TYPE_STATE_REQUEST
	PACKET_TYPE_FRAME_UPDATE
	PACKET_TYPE_FRAME_ACK
	PACKET_TYPE_JOIN
	PACKET_TYPE_JOIN_ACK
	LEN_PACKET_TYPE
)

type Message struct {
	Addr net.Addr
	Body []byte
}

type Packet struct {
	PacketType uint8
	Body       []byte
}

func (p *Packet) UnmarshalBinary(msg []byte) error {
	if len(msg) < 1 {
		return fmt.Errorf("Not enough bytes")
	}
	if uint8(msg[0]) >= LEN_PACKET_TYPE {
		return fmt.Errorf("Not a valid packet type")
	}
	p.PacketType = uint8(msg[0])
	p.Body = make([]byte, len(msg)-1)
	copy(p.Body, msg[1:])
	return nil
}

func (p *Packet) MarshalBinary() ([]byte, error) {
	toret := make([]byte, len(p.Body)+1)
	copy(toret[1:], p.Body)
	toret[0] = byte(p.PacketType)
	return toret, nil
}

type JoinBody struct {
	Username string
}

func (j *JoinBody) UnmarshalBinary(msg []byte) error {
	j.Username = string(msg)
	return nil
}

func (j *JoinBody) MarshalBinary() ([]byte, error) {
	return []byte(j.Username), nil
}

type JoinAckBody struct {
	Frame uint64
	State []byte
}

func (j *JoinAckBody) UnmarshalBinary(msg []byte) error {
	f, bytes := binary.Uvarint(msg)
	if bytes <= 0 {
		return fmt.Errorf("Not a valid uint64")
	}
	j.Frame = f
	j.State = msg[bytes:]
	return nil
}

func (j *JoinAckBody) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(j.State)+8)
	total_bytes := binary.PutUvarint(b, j.Frame)
	copy(b[total_bytes:], j.State)
	return b, nil
}

type FrameUpdateBody struct {
	Frame uint64
	Delta []byte
}

func (f *FrameUpdateBody) UnmarshalBinary(msg []byte) error {
	fr, bytes := binary.Uvarint(msg)
	if bytes <= 0 {
		return fmt.Errorf("Not a valid uint64")
	}
	f.Frame = fr
	f.Delta = msg[bytes:]
	return nil
}

func (f *FrameUpdateBody) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(f.Delta)+8)
	total_bytes := binary.PutUvarint(b, f.Frame)
	copy(b[total_bytes:], f.Delta)
	return b, nil
}

type FrameAckBody struct {
}

type InputBody struct {
	Frame uint64
	Input []byte
}

func (i *InputBody) UnmarshalBinary(msg []byte) error {
	f, bytes := binary.Uvarint(msg)
	if bytes <= 0 {
		return fmt.Errorf("Not a valid uint64")
	}
	i.Frame = f
	i.Input = msg[bytes:]
	return nil
}

func (i *InputBody) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(i.Input)+8)
	total_bytes := binary.PutUvarint(b, i.Frame)
	copy(b[total_bytes:], i.Input)
	return b, nil
}

type StateRequestBody struct {
}
