package mqttbroker

import (
	"fmt"
	"net"
)

//Server is ...
type Server struct {
}

//StartBroker is ...
func (s Server) StartBroker() {
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
}
