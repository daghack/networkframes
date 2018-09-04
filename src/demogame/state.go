package demogame

import (
	"encoding"
	"encoding/binary"
	"gamenet"
)

type State struct {
	counter int64
}

func NewState() *State {
	return &State{}
}

func (s *State) MarshalBinary() ([]byte, error) {
	buffer := make([]byte, 8)
	c := binary.PutVarint(buffer, s.counter)
	return buffer[0:c], nil
}

func (s *State) UnmarshalBinary(msg []byte) error {
	s.counter, _ = binary.Varint(msg)
	return nil
}

type Delta struct {
	Diff int64
}

func (d *Delta) MarshalBinary() ([]byte, error) {
	buffer := make([]byte, 8)
	c := binary.PutVarint(buffer, d.Diff)
	return buffer[0:c], nil
}

func (d *Delta) UnmarshalBinary(msg []byte) error {
	d.Diff, _ = binary.Varint(msg)
	return nil
}

func (s *State) UpdateOnDelta(delta gamenet.Delta) error {
	d := &Delta{}
	err := d.UnmarshalBinary(delta)
	if err != nil {
		return err
	}
	s.counter += d.Diff
	return nil
}

func (s *State) UpdateOnInput(inputs []gamenet.Input) error {
	d := &Delta{}
	for _, input := range inputs {
		err := d.UnmarshalBinary(input)
		if err != nil {
			return err
		}
		s.counter += d.Diff
	}
	return nil
}

func (s *State) GenerateDelta(frame uint64, inputs []gamenet.Input) (encoding.BinaryMarshaler, error) {
	toret := &Delta{}
	d := &Delta{}
	for _, input := range inputs {
		err := d.UnmarshalBinary(input)
		if err != nil {
			return nil, err
		}
		toret.Diff += d.Diff
	}
	return toret, nil
}
