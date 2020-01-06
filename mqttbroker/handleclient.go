package mqttbroker

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	mqtt "github.com/camen6ert/mqtt_parser_go"
)

type client struct {
	id   string
	con  net.Conn
	subs []string
}

//StartRead is ...
func startRead(c net.Conn) {

	for {

		b := make([]byte, 512)

		_, err := c.Read(b)
		log.Printf("% x\n", b)
		if err != nil {
			log.Println(err.Error())
			return
		}

		com, pos, err := mqtt.GetCommand(b)
		fmt.Printf("command: %s\n", com.Command)

		if com.Command == "Connect" {
			_, err := mqtt.HandleConnect(b[pos:])

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			conack := generateConnack()

			c.Write(conack)
		} else if com.Command == "Subscribe" {

			_, err := mqtt.HandleSubscribe(b[pos:])

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			log.Printf("")

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
