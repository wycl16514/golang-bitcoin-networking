One of great creativeness of bitcoin blockchian is it is a distributing system. Thousands of independent nodes can work together just like a
integerated system, and even different nodes may far away from each other, the system can still guarantee no any one can sabotage the whole
system, and the nodes in the system can make sure their data can sycronize with each other and make sure the data integraty.

All these achivements are thanks to the bitcoin network protocl, we will dive deep into the bitcoin networking protocol and make clear how 
such magic is happenning. Following is an example of bitcoin networking package raw data:

```go
f9beb4d976657273696f6e0000000000650000005f1a69d2721101000100000000000000bc8f5e5400000000010000000000000000000000000000000000ffffc61b6409208d010000000000000000000000000000000000ffffcb0071c0208d128035cbc97953f80f2f5361746f7368693a302e392e332fcf05050001
```
Let's deissect it into fields:

1, the first 4 bytes are always the same and referred to as network magic number: f9beb4d9, its usage is to tell receiver that, when you see
these four bytes appear together, then you should konw this is the beginning for bitcoin networking package. And this number use to 
differentiate the mainnet with testnet, for testnet, the four byte is 0b110907.

2， the following 12 bytes is called command: 76657273696f6e0000000000, it used to describe the purpose of this packet. It can be humman
readable string.

3, the following 4 bytes: 65000000 it is payload length in little endian format,

4, the following 4 bytes: 5f1a69d2 is first 4 bytes of  hash256 of the payload.

5, the following bytes are data of the payload

Let's see how to use code to parse the network packet data, first create a new directory named networking and add a new file with name:
network_envelope.go, and add content like following:
```go
package networking

import (
	"bufio"
	"bytes"
	ecc "elliptic_curve"
	"fmt"
	"io"
	"math/big"
	tx "transaction"
)

type NetworkEnvelope struct {
	command []byte
	payload []byte
	testnet bool
	magic   []byte
}

func NewNetworkEnvelope(command []byte, payload []byte, testnet bool) *NetworkEnvelope {
	network := &NetworkEnvelope{
		command: command,
		payload: payload,
		testnet: testnet,
	}

	if testnet {
		network.magic = []byte{0x0b, 0x11, 0x09, 0x07}
	} else {
		network.magic = []byte{0xf9, 0xbe, 0xb4, 0xd9}
	}

	return network
}

func ParseNetwork(rawData []byte, testnet bool) *NetworkEnvelope {
	reader := bytes.NewReader(rawData)
	bufReader := bufio.NewReader(reader)

	magic := make([]byte, 4)
	n, err := io.ReadFull(bufReader, magic)
	if err != nil {
		panic(err)
	}
	if n == 0 {
		panic("connection reset!")
	}

	var expectedMagic []byte

	if testnet {
		expectedMagic = []byte{0x0b, 0x11, 0x09, 0x07}
	} else {
		expectedMagic = []byte{0xf9, 0xbe, 0xb4, 0xd9}
	}
	if bytes.Equal(magic, expectedMagic) != true {
		panic("magic is not right")
	}

	command := make([]byte, 12)
	_, err = io.ReadFull(bufReader, command)
	if err != nil {
		panic(err)
	}

	payloadLenBuf := make([]byte, 4)
	_, err = io.ReadFull(bufReader, payloadLenBuf)
	if err != nil {
		panic(err)
	}
	payloadLen := new(big.Int)
	payloadLen.SetBytes(tx.ReverseByteSlice(payloadLenBuf))
	checksum := make([]byte, 4)
	_, err = io.ReadFull(bufReader, checksum)
	if err != nil {
		panic(err)
	}

	payload := make([]byte, payloadLen.Int64())
	_, err = io.ReadFull(bufReader, payload)
	if err != nil {
		panic(err)
	}

	calculatedChecksum := ecc.Hash256(string(payload))[0:4]
	if !bytes.Equal(checksum, calculatedChecksum) {
		panic("checksum dose not match")
	}

	return NewNetworkEnvelope(command, payload, testnet)
}

func (n *NetworkEnvelope) Serialize() []byte {
	result := make([]byte, 0)
	result = append(result, n.magic...)
	/*
		command need to be 12 bytes long, if it is not enough, we padding
		0x00 at the end
	*/
	command := make([]byte, 0)
	command = append(command, n.command...)
	commandLen := len(command)
	if len(command) < 12 {
		//bug fix, we need to padd command to 12 bytes long
		for i := 0; i < 12-commandLen; i++ {
			command = append(command, 0x00)
		}
	}
	result = append(result, command...)

	payLoadLen := big.NewInt(int64(len(n.payload)))
	result = append(result, tx.ReverseByteSlice(payLoadLen.Bytes())...)
	result = append(result, n.payload...)
	return result
}

func (n *NetworkEnvelope) String() string {
	return fmt.Sprintf("%s : %x\n", string(n.command), n.payload)
}
```
Then we can test the aboved code in main.go as following:
```go
package main

import (
	"encoding/hex"
	"fmt"
	"networking"
)

func main() {
	networkRawData, err := hex.DecodeString("f9beb4d976657273696f6e0000000000650000005f1a69d2721101000100000000000000bc8f5e5400000000010000000000000000000000000000000000ffffc61b6409208d010000000000000000000000000000000000ffffcb0071c0208d128035cbc97953f80f2f5361746f7368693a302e392e332fcf05050001")
	if err != nil {
		panic(err)
	}

	network := networking.ParseNetwork(networkRawData, false)
	fmt.Printf("%s\n", network)
	fmt.Printf("%x\n", network.Serialize())
}
```
Running aboved code will have the following result:
```go
version : 721101000100000000000000bc8f5e5400000000010000000000000000000000000000000000ffffc61b6409208d010000000000000000000000000000000000ffffcb0071c0208d128035cbc97953f80f2f5361746f7368693a302e392e332fcf05050001

f9beb4d976657273696f6e000000000065721101000100000000000000bc8f5e5400000000010000000000000000000000000000000000ffffc61b6409208d010000000000000000000000000000000000ffffcb0071c0208d128035cbc97953f80f2f5361746f7368693a302e392e332fcf05050001
```
The funny thing wo need to notice is that, the command field actually is binary data for string, for example when we convert the command 
field of given network packet header in aboved code, its content is "version". Different command has different payload structure, take 
command "version" as example, following is raw binary payload data for command "version" with ip4 address format(we don't use payload above
because its ip format is ip6 not ip4)
```go
7f110100

0000000000000000

ad17835b00000000

0000000000000000

00000000000000000000ffff00000000

8d20

0000000000000000

00000000000000000000ffff00000000

8d20

f6a8d7a440ec27a1

1b2f70726f6772616d6d696e67626c6f636b636861696e3a302e312f

00000000

01
```

1, the first 4 bytes: 7f110100 in little endian is the protocol version,its value is 70015, at the time of writing, the bitcoin core
version is 70016 .different version number will indicate the node support what kind of bitcoin protocol(BIP XX), 
normally two nodes need to have the same protocol version then they can communicate smoothly.

2, the following 8 bytes in little endian is network services of sender :0000000000000000

3, the following 8 bytes in little edian is timestamp: ad17835b00000000

4, the following 8 bytes in little endian is network services of receiver: 0000000000000000

5, the following 16 bytes is network address of receiver: 00000000000000000000ffff00000000, if the first 12 bytes of this field is 
00000000000000000000ffff, then the ip format is ip4, and the following four bytes is the ip address, in here it is 0.0.0.0

6, the following 2 bytes is network port of receiver: 8d20, it is default to 8d20 which is 8333

8, the followin 8 bytes in little endian is network service of sender: 0000000000000000

9, the following 16 bytes is network address of sender: 00000000000000000000ffff00000000

10, the following 2 bytes is network port of sender: 8d20

11, the following 8 bytes is nonce, used for communicating response: f6a8d7a440ec27a1

12, the following 28 bytes is user agent which is binary data for string: 1b2f70726f6772616d6d696e67626c6f636b636861696e3a302e312f,
at the beginning is the variant int to indicate the raw data length of the string, since the first byte is 1b, which indicates the 
length of raw data is 1b, then the following bytes: 2f70726f6772616d6d696e67626c6f636b636861696e3a302e312f are string data

13, the following 4 bytes is block height: 00000000

14, the last one byte is optional flag for relay based on protocol BIP37: 01

For more details about version command you can check here:

https://en.bitcoin.it/wiki/Protocol_documentation#version

Command "version" is used to initialize connection between two nodes. The payload for "version" command is a simple description the node
itself, such as what kind of protocol version it supports. When the bitcoin node is setup and running, it will find a set of "node seeds"
by using some methods, we will not go into the detail of how to find the "node seeds".

The node seeds is made up by a set of bitcoin full node clients, when the bitcoin node of you begin running, you need to indicate to all 
your neighbors that you are going to joining them, and you send a "hello" message to all your neighbors, this hello message is the network
packet with "version" command.

The payload of the version command package is used to tell neighbors your basic infos such as what kind of bitcoin protocol (BIP) you can
support. Sometimes you may want to say hello to specific peer, then you need to set the receiver ip in the receiver field in the payload.
To be notics, the peer you want to reach may not in the set of your neighbors, then you need to ask you neighbors to relay the message to
its neighbors until the message reach the given peer.

For example, Bob has two neighbors with name Jim and Tom, and Bob want to send a hello message to Alice, but Bon can't reach Alice directly,
therefore he just broadcast the message to Jim and Tom and ask them to help him relay the message to Alice. Jim and Tom will find Alice in
their own neighbors.

If Jim or Tom can find Alice in their neighbors, then they will relay the message to Alice, such relay is the characteristic of p2p network.
When the Alice receives the version message, she will response the message with a "ack" packet and send her own version packet to Bob, and
when Bob receives the version packet from Alice, then he send a "ack" packet to Alice, this process is called handshaking, the whole process
is just like following:

![截屏2024-06-14 13 03 40](https://github.com/wycl16514/golang-bitcoin-networking/assets/7506958/c2365245-d73b-4241-9750-d498ff9f0323)

Let's use code to implement the given handshake process then you will gain a deeper understanding. Create a new file named version_msg.go
and add the following code:

```go
package networking

import (
	"math/big"
	"math/rand"
	"time"
	tx "transaction"
)

type VersionMessage struct {
	command          string
	version          *big.Int
	services         *big.Int
	timestamp        *big.Int
	receiverServices *big.Int
	receiverIP       []byte
	receiverPort     uint16
	senderServices   *big.Int
	senderIP         []byte
	senderPort       uint16
	nonce            []byte
	userAgent        string
	latestBlock      *big.Int
	relay            bool
}

func NewVersionMessage() *VersionMessage {
	/*
		version command used to init handshake between two nodes,
		we only create this command to connect to a local full node
		and we can set all the fields to default value
	*/
	nonceBuf := make([]byte, 8)
	rand.Read(nonceBuf)
	return &VersionMessage{
		command:          "version",
		version:          big.NewInt(70015),
		services:         big.NewInt(0),
		timestamp:        big.NewInt(time.Now().Unix()),
		receiverServices: big.NewInt(0),
		receiverIP:       []byte{0x00, 0x00, 0x00, 0x00},
		receiverPort:     8333,
		senderServices:   big.NewInt(0),
		senderIP:         []byte{0x00, 0x00, 0x00, 0x00},
		senderPort:       8333,
		nonce:            nonceBuf,
		userAgent:        "goloand_bitcoin_lib",
		latestBlock:      big.NewInt(0),
		relay:            false,
	}
}

func (v *VersionMessage) Command() string {
	return v.command
}

func (v *VersionMessage) Serialize() []byte {
	result := make([]byte, 0)
	result = append(result, tx.BigIntToLittleEndian(v.version, tx.LITTLE_ENDIAN_4_BYTES)...)
	result = append(result, tx.BigIntToLittleEndian(v.services, tx.LITTLE_ENDIAN_8_BYTES)...)
	result = append(result, tx.BigIntToLittleEndian(v.timestamp, tx.LITTLE_ENDIAN_8_BYTES)...)
	result = append(result, tx.BigIntToLittleEndian(v.receiverServices, tx.LITTLE_ENDIAN_8_BYTES)...)
	//ip need to be 16 bytes with 0x00...ffff as prefix
	ipBuf := make([]byte, 16)
	for i := 0; i < 12; i++ {
		if i < 10 {
			ipBuf = append(ipBuf, 0x00)
		} else {
			ipBuf = append(ipBuf, 0xff)
		}
	}
	ipBuf = append(ipBuf, v.receiverIP...)
	result = append(result, ipBuf...)

	result = append(result, big.NewInt(int64(v.receiverPort)).Bytes()...)
	result = append(result, tx.BigIntToLittleEndian(v.senderServices, tx.LITTLE_ENDIAN_8_BYTES)...)
	ipBuf = make([]byte, 16)
	for i := 0; i < 12; i++ {
		if i < 10 {
			ipBuf = append(ipBuf, 0x00)
		} else {
			ipBuf = append(ipBuf, 0xff)
		}
	}
	ipBuf = append(ipBuf, v.senderIP...)
	result = append(result, ipBuf...)
	result = append(result, big.NewInt(int64(v.senderPort)).Bytes()...)

	result = append(result, v.nonce...)
	agentLen := tx.EncodeVarint(big.NewInt(int64(len(v.userAgent))))
	result = append(result, agentLen...)
	result = append(result, []byte(v.userAgent)...)
	result = append(result, tx.BigIntToLittleEndian(v.latestBlock, tx.LITTLE_ENDIAN_4_BYTES)...)
	if v.relay {
		result = append(result, 0x01)
	} else {
		result = append(result, 0x00)
	}

	return result

}




```
We can try to run aboved code in main.go as following:
```go
func main() {
	version := networking.NewVersionMessage()
	fmt.Printf("version:%x\n", version.Serialize())
}
```
Then I get the following result, you may have different result compared with me:
```go
version:7f1101df5d6e6600000000000000000000000000000000000000000000000000000000000000000000ffff000000008d2000000000000000000000000000000000000000000000000000000000000000000000ffff000000008d2052fdfc072182654f13676f6c6f616e645f626974636f696e5f6c69620000000000
```

As shown by the imaged aboved, when node A wants to connect with node B, the following process happens:

1, A sends a version message to B

2, B receives the version message from A, if B accepts the version message, it returns a verack message and sends its own version message

3, A receives the version message from B then send a verack message back to B

4, B receives the verack message from A and continues communication

Now let's add the implementation of  verack message. Since its very simple, We can put it in version_msg.go directly as following:

```go
func NewVerAckMessage() *VerAckMessage {
	return &VerAckMessage{
		command: "verack",
	}
}

type VerAckMessage struct {
	command string
}

func (v *VerAckMessage) Serialize() []byte {
	return []byte{}
}
```
Then we use socket to initialize the communication between our node and the bitcoin node we setup aboved, create a new file called
simple_node.go, add the following code:
```go
package networking

import (
	"bytes"
	"fmt"
	"net"
)

type Message interface {
	Command() string
	Serialize() []byte
}

type SimpleNode struct {
	host        string
	port        uint16
	testnet     bool
	receiveMsgs []Message
}

func NewSimpleNode(host string, port uint16, testnet bool) *SimpleNode {
	return &SimpleNode{
		host:    host,
		port:    port,
		testnet: testnet,
	}
}

func (s *SimpleNode) Run() {
	/*
		using socket connect to given host with given port, then
		construct package with payload is version message and send to
		the peer, waiting peer to send back its version message and verack,
		and we send verack back to peer and close the connection
	*/
	conStr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := net.Dial("tcp", conStr)
	if err != nil {
		panic(err)
	}

	s.WaitFor(conn)
}

func (s *SimpleNode) Send(conn net.Conn, msg Message) {
	envelop := NewNetworkEnvelope([]byte(msg.Command()), msg.Serialize(), s.testnet)
	n, err := conn.Write(envelop.Serialize())
	if err != nil {
		panic(err)
	}

	fmt.Printf("write to %d bytes\n", n)
}

func (s *SimpleNode) Read(conn net.Conn) []*NetworkEnvelope {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	/*
		the peer node may return version and verack
		at once
	*/
	var msgs []*NetworkEnvelope
	parsedLen := 0
	for {
		if parsedLen >= n {
			break
		}
		msg := ParseNetwork(buf, s.testnet)
		msgs = append(msgs, msg)
		if parsedLen < n {
			parsedLen += len(msg.Serialize())
			buf = buf[parsedLen:]
		}
	}
	return msgs
}

func (s *SimpleNode) WaitFor(conn net.Conn) {
	s.Send(conn, NewVersionMessage())

	verackReceived := false
	versionReceived := false
	for !verackReceived || !versionReceived {
		msgs := s.Read(conn)
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			command := string(bytes.Trim(msg.command, "\x00"))
			fmt.Printf("command:%s\n", command)
			if command == "verack" {
				fmt.Printf("receiving verack message from peer\n")
				verackReceived = true
			}
			if command == "version" {
				versionReceived = true
				fmt.Printf("receiving version message from peer\n: %s", msg)
				s.Send(conn, NewVerAckMessage())

			}

		}

	}
}

```
Running aboved code we can get the following result:
```go
write to 161 bytes
command:version
receiving version message from peer
: version : 801101000904000000000000403771660000000000000000000000000000000000000000000000000000000000000904000000000000000000000000000000000000000000000000c06b203d5f25c603102f5361746f7368693a32352e302e302f3df20c0000
write to 24 bytes
command:verack
receiving verack message from peer
```
We can see that from the result, our peer sending version and verack packets to us after we sending a version package to it.

