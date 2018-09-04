package gamenet

import (
	"fmt"
	"gamenet/types/packets"
	"net"
)

func NewUDPWriter(pc net.PacketConn) chan<- packets.Message {
	newchan := make(chan packets.Message, 128)
	go func(conn net.PacketConn, readchan <-chan packets.Message) {
		for packet := range readchan {
			_, err := conn.WriteTo(packet.Body, packet.Addr)
			if err != nil {
				fmt.Println(err)
			}
		}
	}(pc, newchan)
	return newchan
}
