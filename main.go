package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	mqtt "github.com/camen6ert/mqtt_parser_go"
)

func main() {
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:1883")
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go startRead(conn)
		}
	}()

	select {}
}

func startRead(c net.Conn) {

	for {

		b := make([]byte, 512)

		_, err := c.Read(b)
		fmt.Printf("% x\n", b)
		//fmt.Println(k)
		if err != nil {
			//if err != io.EOF {
			fmt.Println(err.Error())
			return
			//}
		}

		com, pos, err := mqtt.GetCommand(b)
		fmt.Printf("command: %s\n", com.Command)
		if com.Command == "Connect" {
			connect, err := mqtt.HandleConnect(b[pos:])

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			fmt.Println("%s\n", connect)
		}
	}
}

func generateConnack() []byte {

	bs := make([]byte, 5)

	bs[0] = 0x20
	bs[1] = 0x03
	bs[2] = 0x00
	bs[3] = 0x00

	return bs

}

func generateSuback(packID uint16) []byte {

	bs := make([]byte, 7)

	pi := make([]byte, 2)
	binary.BigEndian.PutUint16(pi, uint16(packID))

	bs[0] = 0x90
	bs[1] = 0x04
	bs[2] = pi[0]
	bs[3] = pi[1]
	bs[4] = 0x00
	bs[5] = 0x00
	bs[6] = 0x87

	return bs
}
