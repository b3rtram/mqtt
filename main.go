package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	// Time to wait before starting closing clients when in LD mode.
	connect    = 0x10
	publish    = 0x30
	subscribe  = 0x82
	suback     = 0x09
	disconnect = 0xe0
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

		p := make([]byte, 512)

		_, err := c.Read(p)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("data: % x\n", p)
		mqttLen, read := getVarByteInt(p[1:])
		fmt.Printf("%d\n", mqttLen)

		pos := read + 1

		switch p[0] {
		case connect:
			handleConnect(p[pos : mqttLen+pos])

			c.Write(generateConnack())

			if err != nil {
				fmt.Printf("write connack %s\n", err.Error())
			}

		case publish:
			topic, len := getUtf8(p[pos:])
			fmt.Printf("topic: %s %d\n", topic, len)
			pos += len
			payLen := mqttLen - pos + 2

			msg := p[pos : pos+payLen]
			fmt.Printf("pub payload: %s\n", string(msg))

		case subscribe:

			packetID := getUint16(p[pos], p[pos+1])
			fmt.Printf("packetID %d\n", packetID)
			pos += 2
			propLen := p[pos]
			pos++
			subID := 0

			for i := 0; i < int(propLen); i++ {

				b := p[pos+i]

				switch int(b) {
				case 0x0b:
					var r int
					subID, r = getVarByteInt(p)
					fmt.Printf("subID %d %d\n", subID, r)
					pos += r
				case 0x26:
					user, r := getUtf8(p)
					fmt.Printf("user: %s\n", user)
					pos += r
				}
			}

			for {
				topic, r := getUtf8(p[pos:])
				fmt.Printf("topic: %s %d\n", topic, r)
				pos += r + 1

				if pos > mqttLen {
					break
				}
			}

			c.Write(generateSuback(subID))

		case disconnect:
			fmt.Println("Disconnect")
		}

	}
}

func handleConnect(vhead []byte) {

	fmt.Printf("%s", vhead)

	//check variable header for correctness
	if vhead[0] != 0x00 || vhead[1] != 0x04 || vhead[2] != 0x4d || vhead[3] != 0x51 || vhead[4] != 0x54 || vhead[5] != 0x54 {
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
		case 0x11:
			sessionExp := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])
			i += 4
			fmt.Printf("session expires %d", sessionExp)

		case 0x21:
			//Receive Maximum
			receiveMax := getUint16(vhead[i+1], vhead[i+2])
			i += 2
			fmt.Printf("receive maximum %d", receiveMax)

		case 0x27:
			//Maximum Packet Size
			maxPacketSize := getUint32(vhead[i+1], vhead[i+2], vhead[i+3], vhead[i+4])
			i += 4
			fmt.Printf("maximum packet size %d", maxPacketSize)

		case 0x22:
			//Topic Alias Maximum
			receiveMax := getUint16(vhead[i+1], vhead[i+2])
			i += 4
			fmt.Printf("receive maximum %d", receiveMax)

		case 0x19:
			//Request response info
			reqResInfo := int(vhead[i+1])
			i++
			fmt.Printf("Request response information %d", reqResInfo)

		case 0x17:
			//Request Problem Information
			reqProbInfo := int(vhead[i+1])
			i++
			fmt.Printf("Request Problem Information %d", reqProbInfo)

		case 0x26:
			unBuf := make([]byte, 2)
			unBuf[0] = vhead[i+1]
			unBuf[1] = vhead[i+2]
			username := string(unBuf)

			fmt.Printf("Username %s", username)
		case 0x15:
			//Authentication Extension
			fmt.Printf("Authentication Extension")
		case 0x16:
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

}

func generateConnack() []byte {

	bs := make([]byte, 5)

	bs[0] = 0x20
	bs[1] = 0x03
	bs[2] = 0x00
	bs[3] = 0x00

	return bs

}

func generateSuback(packID int) []byte {

	bs := make([]byte, 6)

	pi := make([]byte, 2)
	binary.LittleEndian.PutUint16(pi, uint16(packID))

	bs[0] = 0x90
	bs[1] = 0x04
	bs[2] = pi[0]
	bs[3] = pi[1]
	bs[4] = 0x00
	bs[5] = 0x00

	return bs
}

func getVarByteInt(bs []byte) (int, int) {
	multiplier := 1
	value := 0
	a := 0
	for {
		encodedByte := bs[a]
		value += (int(encodedByte) & 127) * multiplier
		if multiplier > 128*128*128 {
			break
		}

		multiplier *= 128
		a++
		if (encodedByte & 128) == 0 {
			break
		}

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
