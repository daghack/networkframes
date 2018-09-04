package gamenet

import (
	"encoding"
	"fmt"
	"gamenet/types/packets"
	"net"
)

type GameClient struct {
	Username      string
	ServerAddr    *net.UDPAddr
	current_frame uint64
	current_state State
	udp_listener  <-chan packets.Message
	udp_writer    chan<- packets.Message
	joined        bool
}

func NewGameClient(username, hostaddr, serveraddr string, state State) (*GameClient, error) {
	conn, err := net.ListenPacket("udp", hostaddr)
	if err != nil {
		return nil, err
	}
	saddr, err := net.ResolveUDPAddr("udp", serveraddr)
	if err != nil {
		return nil, err
	}
	return &GameClient{
		Username:      username,
		udp_listener:  NewUDPListener(conn, 1024),
		udp_writer:    NewUDPWriter(conn),
		current_state: state,
		ServerAddr:    saddr,
	}, nil
}

func (client *GameClient) FrameAck(frame uint64, delta []byte) error {
	return nil
}

func (client *GameClient) Join() error {
	jb := &packets.JoinBody{
		Username: client.Username,
	}
	b, err := jb.MarshalBinary()
	if err != nil {
		return err
	}
	packet := packets.Packet{
		PacketType: packets.PACKET_TYPE_JOIN,
		Body:       b,
	}
	packetbytes, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	client.udp_writer <- packets.Message{
		Body: packetbytes,
		Addr: client.ServerAddr,
	}
	return nil
}

func (client *GameClient) Input(input encoding.BinaryMarshaler) error {
	if !client.joined {
		return fmt.Errorf("Not yet joined")
	}
	ibytes, err := input.MarshalBinary()
	if err != nil {
		return err
	}
	ib := &packets.InputBody{
		Frame: client.current_frame,
		Input: ibytes,
	}
	b, err := ib.MarshalBinary()
	if err != nil {
		return err
	}
	packet := packets.Packet{
		PacketType: packets.PACKET_TYPE_INPUT,
		Body:       b,
	}
	packetbytes, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	client.udp_writer <- packets.Message{
		Body: packetbytes,
		Addr: client.ServerAddr,
	}
	return nil
}

func (client *GameClient) RunGameClient() error {
	packet := &packets.Packet{}
	frameUpdate := &packets.FrameUpdateBody{}
	joinAck := &packets.JoinAckBody{}
	var err error
	for msg := range client.udp_listener {
		err = packet.UnmarshalBinary(msg.Body)
		if err != nil {
			return err
		}
		switch packet.PacketType {
		case packets.PACKET_TYPE_JOIN_ACK:
			err = joinAck.UnmarshalBinary(packet.Body)
			if err != nil {
				return err
			}
			fmt.Printf("Received Join Ack, Frame %d\n", joinAck.Frame)
			client.current_frame = joinAck.Frame
			client.current_state.UnmarshalBinary(joinAck.State)
			fmt.Printf("State: %v\n", client.current_state)
			client.joined = true
		case packets.PACKET_TYPE_FRAME_UPDATE:
			if !client.joined {
				continue
			}
			fmt.Println(packet.Body)
			err = frameUpdate.UnmarshalBinary(packet.Body)
			if err != nil {
				return err
			}
			err := client.current_state.UpdateOnDelta(frameUpdate.Delta)
			if err != nil {
				return err
			}
			client.current_frame = frameUpdate.Frame
			client.FrameAck(frameUpdate.Frame, frameUpdate.Delta)
			fmt.Printf(
				"Frame %d, State %v\n",
				client.current_frame,
				client.current_state,
			)
		}
	}
	return fmt.Errorf("UDP Listener Chan Closed")
}
