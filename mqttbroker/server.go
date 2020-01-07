package mqttbroker

import (
	"fmt"
	"net"

	mqtt "github.com/camen6ert/mqtt_parser_go"
)

//Server is ...
type Server struct {
	clients []*client
	chans   channels
}

type channels struct {
	//channels
	addClientChan   chan client
	removeClienChan chan client
	subscribeChan   chan mqtt.Subscribe
	publishChan     chan mqtt.Publish
}

//StartBroker is ...
func (s Server) StartBroker() {

	s.chans = channels{}

	s.chans.addClientChan = make(chan client)
	s.chans.removeClienChan = make(chan client)
	s.chans.subscribeChan = make(chan mqtt.Subscribe)
	s.chans.publishChan = make(chan mqtt.Publish)

	go s.addClient()
	go s.removeClient()
	go s.subscribe()
	go s.publish()

	listener, err := net.Listen("tcp", "127.0.0.1:1883")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go startRead(conn, s.chans)
	}
}

func (s Server) addClient(c chan client) {

}

func (s Server) removeClient(c chan client) {

}

func (s Server) subscribe(c chan mqtt.Subscribe) {

}

func (s Server) publish(c chan mqtt.Publish) {

}
