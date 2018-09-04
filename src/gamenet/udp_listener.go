package gamenet

import (
	"fmt"
	"gamenet/types/packets"
	"net"
)

func NewUDPListener(pc net.PacketConn, buffersize int) <-chan packets.Message {
	newchan := make(chan packets.Message, 128)
	go func(conn net.PacketConn, writechan chan<- packets.Message) {
		buffer := make([]byte, buffersize)
		for {
			bytes, addr, err := conn.ReadFrom(buffer)
			body := make([]byte, bytes)
			copy(body, buffer)
			if err != nil {
				fmt.Println(err)
			}
			writechan <- packets.Message{
				Addr: addr,
				Body: body,
			}
		}
	}(pc, newchan)
	return newchan
}
