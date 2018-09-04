package gamenet

import (
	"encoding"
	"fmt"
	"gamenet/types/packets"
	"net"
	"time"
)

type Input []byte
type Delta []byte

type State interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	UpdateOnDelta(Delta) error
	UpdateOnInput([]Input) error
	GenerateDelta(uint64, []Input) (encoding.BinaryMarshaler, error)
}

type Client struct {
	Username  string
	LastFrame uint64
	Addr      net.Addr
}

type GameServer struct {
	step_timer    <-chan time.Time
	udp_listener  <-chan packets.Message
	udp_writer    chan<- packets.Message
	current_frame uint64
	current_state State
	frame_buffer  [128][]Input
	clients       map[string]Client
}

func NewGameServer(state State) (*GameServer, error) {
	var frame_buffer [128][]Input
	for i := range frame_buffer {
		frame_buffer[i] = []Input{}
	}
	conn, err := net.ListenPacket("udp", "localhost:7788")
	if err != nil {
		return nil, err
	}
	return &GameServer{
		step_timer:    time.Tick(16 * time.Millisecond),
		udp_listener:  NewUDPListener(conn, 1024),
		udp_writer:    NewUDPWriter(conn),
		current_state: state,
		frame_buffer:  frame_buffer,
		clients:       map[string]Client{},
	}, nil
}

func (g *GameServer) RequestDeltaSinceFrame(frame uint64) (Delta, error) {
	if frame < g.current_frame {
		return nil, fmt.Errorf("Outdated Frame")
	} else if frame > g.current_frame+uint64(len(g.frame_buffer)) {
		return nil, fmt.Errorf("Non-buffered Frame Requested")
	}
	input_list := []Input{}
	for i := uint64(0); i <= frame-g.current_frame; i += 1 {
		index := (i + g.current_frame) % uint64(len(g.frame_buffer))
		input_list = append(input_list, g.frame_buffer[index]...)
	}
	marshaler, err := g.current_state.GenerateDelta(frame, input_list)
	if err != nil {
		return nil, err
	}
	return marshaler.MarshalBinary()
}

func (g *GameServer) HandleInput(frame uint64, input Input) error {
	if frame < g.current_frame {
		return fmt.Errorf("Outdated Frame")
	} else if frame > g.current_frame+uint64(len(g.frame_buffer)) {
		return fmt.Errorf("Non-buffered Frame Requested")
	}
	index := frame % uint64(len(g.frame_buffer))
	g.frame_buffer[index] = append(g.frame_buffer[index], input)
	return nil
}

func (g *GameServer) GameServerStep() error {
	index := g.current_frame % uint64(len(g.frame_buffer))
	err := g.current_state.UpdateOnInput(g.frame_buffer[index])
	if err != nil {
		return err
	}
	marshaler, err := g.current_state.GenerateDelta(g.current_frame, g.frame_buffer[index])
	if err != nil {
		return err
	}
	delta, err := marshaler.MarshalBinary()
	if err != nil {
		return err
	}
	update := packets.FrameUpdateBody{
		Frame: g.current_frame + 1,
		Delta: delta,
	}
	if err != nil {
		return err
	}
	body, err := update.MarshalBinary()
	g.frame_buffer[index] = []Input{}
	g.current_frame += 1
	tosend := packets.Packet{
		PacketType: packets.PACKET_TYPE_FRAME_UPDATE,
		Body:       body,
	}
	for _, client := range g.clients {
		data, err := tosend.MarshalBinary()
		if err != nil {
			return err
		}
		g.udp_writer <- packets.Message{
			Body: data,
			Addr: client.Addr,
		}
	}
	return nil
}

func (g *GameServer) JoinAck(addr net.Addr) error {
	state, err := g.current_state.MarshalBinary()
	if err != nil {
		return err
	}
	joinack := packets.JoinAckBody{
		Frame: g.current_frame,
		State: state,
	}
	body, err := joinack.MarshalBinary()
	if err != nil {
		return err
	}
	packet := packets.Packet{
		PacketType: packets.PACKET_TYPE_JOIN_ACK,
		Body:       body,
	}
	msg, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	g.udp_writer <- packets.Message{
		Body: msg,
		Addr: addr,
	}
	return nil
}

func (g *GameServer) RunGameServer() error {
	var err error
	packet := &packets.Packet{}
	join := &packets.JoinBody{}
	input := &packets.InputBody{}
	for {
		select {
		case <-g.step_timer:
			err = g.GameServerStep()
			if err != nil {
				return err
			}
		case msg := <-g.udp_listener:
			err = packet.UnmarshalBinary(msg.Body)
			if err != nil {
				return err
			}
			switch packet.PacketType {
			case packets.PACKET_TYPE_JOIN:
				fmt.Printf("Received Join From: %v\n", msg.Addr)
				err = join.UnmarshalBinary(packet.Body)
				if err != nil {
					return err
				}
				fmt.Printf("Username: %s\n", join.Username)
				g.clients[join.Username] = Client{
					Username:  join.Username,
					Addr:      msg.Addr,
					LastFrame: g.current_frame,
				}
				g.JoinAck(msg.Addr)
			case packets.PACKET_TYPE_INPUT:
				fmt.Printf("Received Input From: %v\n", msg.Addr)
				err = input.UnmarshalBinary(packet.Body)
				if err != nil {
					return err
				}
				fmt.Printf("Frame %d, Input %v\n", input.Frame, input.Input)
				g.HandleInput(input.Frame, input.Input)
			case packets.PACKET_TYPE_STATE_REQUEST:
				fmt.Printf("Received State Request From: %v\n", msg.Addr)
			case packets.PACKET_TYPE_FRAME_ACK:
				fmt.Printf("Received Frame Ack From: %v\n", msg.Addr)
			}
			fmt.Printf("Current Frame: %d\n", g.current_frame)
		}
	}
}
