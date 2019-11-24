package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	// Time to wait before starting closing clients when in LD mode.
	connect = 16
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

	header := make([]byte, 2)

	_, err := c.Read(header)
	if err != nil {
		fmt.Println(err.Error())
	}

	len := header[1]

	//Handle Connect
	if header[0] == connect {
		vhead := make([]byte, len)
		_, err = c.Read(vhead)
		if err != nil {
			fmt.Println(err.Error())
		}

		//check variable header for correctness
		if vhead[0] != 0 || vhead[1] != 4 || vhead[2] != 77 || vhead[3] != 81 || vhead[4] != 84 || vhead[5] != 84 {
			fmt.Println("error in CONNECT")
		}

		ver := vhead[6]
		if ver != 5 {
			fmt.Println("error < 5 not supported")
		}

		//Check UserName Flag
		un := false
		if vhead[7]&0x80 == 0x80 {
			un = true
		}

		fmt.Printf("username %t\n", un)

		//Check Password flag
		pwd := false
		if vhead[7]&0x40 == 0x40 {
			pwd = true
		}

		fmt.Printf("pasword %t\n", pwd)

		//Check Will Retain flag
		willRetain := false
		if vhead[7]&0x20 == 0x20 {
			willRetain = true
		}

		fmt.Printf("willRetain %t\n", willRetain)

		//Check will QOS
		qos := vhead[7]&0x10 + vhead[7]&0x08
		fmt.Printf("will qos %d", qos)

		//Check will flag
		willFlag := false
		if vhead[7]&0x04 == 0x04 {
			willFlag = true
		}

		fmt.Printf("will %t\n", willFlag)

		//Check clean start
		cleanStart := false
		if vhead[7]&0x02 == 0x02 {
			cleanStart = true
		}

		fmt.Printf("clean start %t\n", cleanStart)

		if vhead[7]&0x01 == 0x01 {
			fmt.Println("ERROR")
		}

		keepAliveBuf := make([]byte, 2)
		keepAliveBuf[0] = vhead[8]
		keepAliveBuf[1] = vhead[9]

		//fmt.Printf("% 08b", vhead[8])

		keepAlive := binary.BigEndian.Uint16(keepAliveBuf)
		fmt.Printf("%d\n", keepAlive)

		propLen := vhead[10]
		fmt.Printf("properties len %d", propLen)

		for i := 11; i < int(propLen)+11; i++ {
			b := vhead[i]

			switch int(b) {
			//Session Expiry Interval
			case 17:
				sessionExp := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])

				fmt.Printf("session expires %d", sessionExp)

			case 33:
				//Receive Maximum
				receiveMax := getUint16(vhead[i+1], vhead[i+2])
				fmt.Printf("receive maximum %d", receiveMax)

			case 39:
				//Maximum Packet Size
				maxPacketSize := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])

				fmt.Printf("maximum packet size %d", maxPacketSize)

			case 34:
				//Topic Alias Maximum
				receiveMax := getUint16(vhead[i+1], vhead[i+2])
				fmt.Printf("receive maximum %d", receiveMax)

			default:
				fmt.Println("ERROR")

			}

		}
	}

}

func getUint16(b1 byte, b2 byte) uint16 {

	bs := make([]byte, 2)
	bs[0] = b1
	bs[1] = b2

	return binary.BigEndian.Uint16(bs)
}

func getUint32(b1 byte, b2 byte, b3 byte, b4 byte) uint32 {

	bs := make([]byte, 4)
	bs[0] = b1
	bs[1] = b2
	bs[2] = b3
	bs[3] = b4

	return binary.BigEndian.Uint32(bs)
}
