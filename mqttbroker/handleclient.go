package mqttbroker

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"

	mqtt "github.com/camen6ert/mqtt_parser_go"
)

type client struct {
	id  string
	con net.Conn
}

type subscribe struct {
	client    client
	subscribe mqtt.Subscribe
	//on subscription get publish messages over that channel
	pubchan chan publish
}

type publish struct {
	client  client
	publish mqtt.Publish
}

//StartRead is ...
func startClient(c net.Conn, s channels) {
	client := client{con: c}
	pubchan := make(chan publish)

	go handlePub(pubchan, &client)

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
			connect, err := mqtt.HandleConnect(b[pos:])

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			client.id = connect.ClientID
			s.addClientChan <- client

			conack := generateConnack()
			c.Write(conack)

		} else if com.Command == "Subscribe" {

			sub, err := mqtt.HandleSubscribe(b[pos:], com.MqttLen)

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			subscribe := subscribe{client: client, subscribe: sub, pubchan: pubchan}
			s.subscribeChan <- subscribe

			suback := generateSuback(sub.PacketID)
			c.Write(suback)

		} else if com.Command == "Publish" {

			p, err := mqtt.HandlePublish(b, pos, com.MqttLen)

			if err != nil {
				log.Fatalf("%s \n", err.Error())
			}

			publish := publish{client: client, publish: p}
			s.publishChan <- publish

		} else if com.Command == "PingReq" {

			c.Write(generatePingresp())

			log.Printf("resend ping response")
		}

	}
}

func handlePub(p chan publish, c *client) {
	for {
		pub := <-p
		client := *c
		n, err := client.con.Write(pub.publish.CompleteMsg)

		log.Printf("%d %s\n", n, err.Error())
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

func generatePingresp() []byte {
	bs := make([]byte, 5)

	bs[0] = 0x13
	bs[1] = 0x00

	return bs
}
