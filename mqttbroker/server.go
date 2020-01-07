package mqttbroker

import (
	"fmt"
	"log"
	"net"
	//mqtt "github.com/camen6ert/mqtt_parser_go"
)

//Server is ...
type Server struct {
	clients       []client
	chans         channels
	subscriptions []subscribe
}

type channels struct {
	//channels
	addClientChan    chan client
	removeClientChan chan client
	subscribeChan    chan subscribe
	publishChan      chan publish
}

//StartBroker is ...
func (s Server) StartBroker() {

	s.chans = channels{}

	s.chans.addClientChan = make(chan client)
	s.chans.removeClientChan = make(chan client)
	s.chans.subscribeChan = make(chan subscribe)
	s.chans.publishChan = make(chan publish)

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

		go startClient(conn, s.chans)
	}
}

func (s *Server) addClient() {
	for {
		client := <-s.chans.addClientChan
		s.clients = append(s.clients, client)
	}
}

func (s *Server) removeClient() {

}

func (s *Server) subscribe() {
	for {
		sub := <-s.chans.subscribeChan
		log.Printf("%s\n", sub.subscribe.Topic[0])

		s.subscriptions = append(s.subscriptions, sub)
	}
}

func (s *Server) publish() {
	for {
		pub := <-s.chans.publishChan
		log.Printf("%s\n", pub.publish.Topic)

		for _, e := range s.subscriptions {
			if e.subscribe.Topic[1] == pub.publish.Topic {
				pub.client = e.client
				e.pubchan <- pub
			}
		}
	}
}
