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
	fmt.Println(len)
	c.

	//Handle Connect
	if header[0] == connect {
		vhead := make([]byte, len)
		_, err = c.Read(vhead)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Printf("%d", vhead)

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

		keepAlive := getUint16(vhead[8], vhead[9])
		fmt.Printf("%d\n", keepAlive)

		propLen := vhead[10]
		fmt.Printf("properties len %d", propLen)

		//Scan properties
		m := 11
		for i := m; i < int(propLen)+m; i++ {
			b := vhead[i]

			switch int(b) {
			//Session Expiry Interval
			case 17:
				sessionExp := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])
				i += 4
				fmt.Printf("session expires %d", sessionExp)

			case 33:
				//Receive Maximum
				receiveMax := getUint16(vhead[i+1], vhead[i+2])
				i += 2
				fmt.Printf("receive maximum %d", receiveMax)

			case 39:
				//Maximum Packet Size
				maxPacketSize := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])
				i += 4
				fmt.Printf("maximum packet size %d", maxPacketSize)

			case 34:
				//Topic Alias Maximum
				receiveMax := getUint16(vhead[i+1], vhead[i+2])
				i += 4
				fmt.Printf("receive maximum %d", receiveMax)

			case 25:
				//Request response info
				reqResInfo := int(vhead[i+1])
				i++
				fmt.Printf("Request response information %d", reqResInfo)

			case 23:
				//Request Problem Information
				reqProbInfo := int(vhead[i+1])
				i++
				fmt.Printf("Request Problem Information %d", reqProbInfo)

			case 38:
				unBuf := make([]byte, 2)
				unBuf[0] = vhead[i+1]
				unBuf[1] = vhead[i+2]
				username := string(unBuf)

				fmt.Printf("Username %s", username)
			case 21:
				//Authentication Extension
				fmt.Printf("Authentication Extension")
			case 22:
				//Authentication Data
				fmt.Printf("Authentication Data")
			default:
				fmt.Println("ERROR")

			}

			m = i
		}

		clientID, a := getUtf8(vhead[m:])
		m += a
		fmt.Printf("ClientID %s %d\n", clientID, a)

		if willFlag == true {

		}

		if un == true {
			username, a := getUtf8(vhead[m:])
			fmt.Printf("username %s", username)
			m += a
		}

		if pwd == true {
			password, _ := getUtf8(vhead[m:])
			fmt.Printf("username %s", password)
		}

		c.Write(generateConnack())
	}

}

func generateConnack() []byte {

	bs := make([]byte, 4)

	bs[0] = 0x10
	bs[1] = 0x04
	bs[2] = 0x00
	bs[3] = 0x00

	return bs

}

func getVarByteInt(bs []byte) (int, int) {
	multiplier := 1
	value := 0
	a := 0

	encodedByte := bs[a]
	fmt.Printf("%b\n", encodedByte)
	value += int(encodedByte&127) * multiplier
	if multiplier > 128*128*128 {
		fmt.Printf("ERROR multiplier")
	}

	multiplier *= 128
	a++
	encodedByte = bs[a]

	for (encodedByte & 128) != 0 {

		value += int(encodedByte&127) * multiplier
		if multiplier > 128*128*128 {
			fmt.Printf("ERROR multiplier")
		}

		multiplier *= 128
		a++
		encodedByte = bs[a]
	}

	return value, a
}

func getUtf8(bs []byte) (string, int) {
	len := getUint16(bs[0], bs[1])

	clientID := make([]byte, len)
	a := 0
	for i := 0; i < int(len); i++ {
		clientID[i] = bs[i+2]
		a++
	}

	return string(clientID), a + 2
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
