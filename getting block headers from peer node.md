In previous section, we successfully setup the communication channel with our peer node.Then we can use the channel to send more requests to
the peer to ask more data. The most request data is requesting block header info from a full node. When you have block headers then you can
request the block body from multiple peers or doing many useful works by those headers.

As we have seen the version command before, the command for getting block headers has name "getheaders" and following is raw binary data for
payload of getheaders command:
```go
7f11010001a35bd0ca2f4a88c4eda6d213e2378a5758dfcd6af437120000000000000000000000000000000000000000000000000000000000000000000000000000000000
```
Let's dissect the aboved data chunk into fields as following:

1, the first four bytes is protocol version as we have seen before: 7f110100

2, the following bytes is varint used to indicate the number of hash, here the value is 01 which is only one byte

3, the following bytes are the starting block hash we want to request:

a35bd0ca2f4a88c4eda6d213e2378a5758dfcd6af43712000000000000000000

4, the following bytes are the eding block hash we want to request:

0000000000000000000000000000000000000000000000000000000000000000

the all zeros here means we want to get as block headers as many as possible but the maximum number we can get is 2000, which is almost
a difficulty adjustment period(2016 blocks)

Let's see how we can construct the payload for the command getheaders, we create a new file named get_headers.go and add the following code:
```go
package networking

import (
	"encoding/hex"
	"math/big"
	tx "transaction"
)

type GetHeaderMessage struct {
	command    string
	version    *big.Int
	numHashes  *big.Int
	startBlock []byte
	endBlock   []byte
}

func GetGenesisBlockHash() []byte {
	genesisBlockRawData, err := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a29ab5f49ffff001d1dac2b7c")
	if err != nil {
		panic(err)
	}
	genesisBlock := tx.ParseBlock(genesisBlockRawData)
	return genesisBlock.Hash()
}

func NewGetHeaderMessage(startBlock []byte) *GetHeaderMessage {
	return &GetHeaderMessage{
		command:    "getheaders",
		version:    big.NewInt(70015),
		numHashes:  big.NewInt(1),
		startBlock: startBlock,
		endBlock:   make([]byte, 32),
	}
}


func (g *GetHeaderMessage) Serialize() []byte {
	result := make([]byte, 0)
	result = append(result, tx.BigIntToLittleEndian(g.version,
		tx.LITTLE_ENDIAN_4_BYTES)...)
	result = append(result, tx.EncodeVarint(g.numHashes)...)
	result = append(result, tx.ReverseByteSlice(g.startBlock)...)
	result = append(result, tx.ReverseByteSlice(g.endBlock)...)
	return result
}

```
Then we can test the aboved code in main.go as following:
```go
func main() {
	getHeadersMsg := networking.NewGetHeaderMessage(networking.GetGenesisBlockHash())
	fmt.Printf("raw data for get headers msg:%x\n", getHeadersMsg.Serialize())
}
```
The result for running aboved code is :
```go
raw data for get headers msg:7f110100016fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d61900000000000000000000000000000000000000000000000000000000000000000000000000
```
As we have seen the handshaking process in previous section. After the handshaking, two nodes setup the communication channel, then one node
can send request to other node. We can try to send get header request to our fullnode peer after the handshaking process. When the full node
peer receives the get header request, it will return infos about given block headers, and we can parse those block headers as we have done
in previous section.

Let's see an example of response raw data for the get header request:
```go
0200000020df3b053dc46f162a9b00c7f0d5124e2676d47bbe7c5d0793a500000000000000ef445fef2ed495c275892206ca533e7411907971013ab83e3b47bd0d692d14d4dc7c835b67d8001ac157e670000000002030eb2540c41025690160a1014c577061596e32e426b712c7ca00000000000000768b89f07044e6130ead292a3f51951adbd2202df447d98789339937fd006bd44880835b67d8001ade09204600
```

1, the first several bytes used to indicate the number of block headers in the data, it is a variant int type, in the example aboved,
it is only one byte long and its value is 0x02, which means there are two block headers returned in the data

2, the following 80 bytes are the raw data for the first block header:
00000020df3b053dc46f162a9b00c7f0d5124e2676d47bbe7c5d0793a500000000000000ef445fef2ed495c275892206ca533e7411907971013ab83e3b47bd0d692d14d4dc7c835b67d8001ac157e67

3, the following bytes in varint int is the number of transactions for the given block, it is always has value 0x00

4, the following 80 bytes are the second block header:
00000002030eb2540c41025690160a1014c577061596e32e426b712c7ca00000000000000768b89f07044e6130ead292a3f51951adbd2202df447d98789339937fd006bd44880835b67d8001ade092046

5, the final bytes in variant int is the transaction numbe for the second block, it is always has value 0x00

Let's see how to use code to parse the return data for get header request:
```go
func LenOfVarint(val *big.Int) int {
	shiftBytes := len(val.Bytes())
	if val.Cmp(big.NewInt(0xfd)) > 0 {
		//if the value bigger than 0xfd, we need to shift
		//one more byte
		shiftBytes += 1
	}

	return shiftBytes
}

func ParseGetHeader(rawData []byte) []*tx.Block {
	reader := bytes.NewReader(rawData)
	bufReader := bufio.NewReader(reader)
	numHeaders := tx.ReadVarint(bufReader)
	fmt.Printf("header count:%d\n", numHeaders.Int64())
	shiftBytes := LenOfVarint(numHeaders)
	rawData = rawData[shiftBytes:]
	blocks := make([]*tx.Block, 0)

	for i := 0; i < int(numHeaders.Int64()); i++ {
		block := tx.ParseBlock(rawData)
		blocks = append(blocks, block)

		rawData = rawData[len(block.Serialize()):]
		reader := bytes.NewReader(rawData)
		bufReader := bufio.NewReader(reader)
		numTxs := tx.ReadVarint(bufReader)
		if numTxs.Cmp(big.NewInt(0)) != 0 {
			//number of transaction should be 0
			panic("number of transaction is not 0")
		}

		shift := LenOfVarint(numTxs)
		if shift == 0 {
			/*
				big.Int will trim any prefix 0x00 in the buf,
				therefore if the buf contains only one byte
				with value 0x00, then Bytes() will return an
				empty slice
			*/
			shift = 1
		}
		rawData = rawData[shift:]
	}
	return blocks
}
```
Then we can test the aboved code like following:
```go
func main() {

getHeaderRaw, err := hex.DecodeString("0200000020df3b053dc46f162a9b00c7f0d5124e2676d47bbe7c5d0793a500000000000000ef445fef2ed495c275892206ca533e7411907971013ab83e3b47bd0d692d14d4dc7c835b67d8001ac157e670000000002030eb2540c41025690160a1014c577061596e32e426b712c7ca00000000000000768b89f07044e6130ead292a3f51951adbd2202df447d98789339937fd006bd44880835b67d8001ade09204600")
	if err != nil {
		panic(err)
	}
	blocks := networking.ParseGetHeader(getHeaderRaw)
	for i := 0; i < len(blocks); i++ {
		fmt.Printf("%s\n", blocks[i])
	}
}

```
Running aboved code will get the following result:
```go
header count:2
version:20000000
previous block id:00000000000000a593075d7cbe7bd476264e12d5f0c7009b2a166fc43d053bdf
merkle root:d4142d690dbd473b3eb83a0171799011743e53ca06228975c295d42eef5f44ef
time stamp:5b837cdc
bits:67d8001a
nonce:c157e670
hash:00000000000000cac712b726e4326e596170574c01a16001692510c44025eb30

version:20000000
previous block id:00000000000000cac712b726e4326e596170574c01a16001692510c44025eb30
merkle root:d46b00fd3799338987d947f42d20d2db1a95513f2a29ad0e13e64470f0898b76
time stamp:5b838048
bits:67d8001a
nonce:de092046
hash:00000000000000beb88910c46f6b442312361c6693a7fb52065b583979844910
```
Now we have the capability to parse block header, then we can send getheader command to the full node peer and ask it to return given number
of block headers back to us, here is the code change we add in the simple node:
```go
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

	s.GetHeaders(conn)
}

func (s *SimpleNode) GetHeaders(conn net.Conn) {
	//after handshaking we send get header request
	getHeadersMsg := NewGetHeaderMessage(GetGenesisBlockHash())
	fmt.Printf("get header raw:%x\n", getHeadersMsg.Serialize())

	s.Send(conn, getHeadersMsg)
	receivedGetHeader := false
	for !receivedGetHeader {
		msgs := s.Read(conn)
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			fmt.Printf("receiving command: %s\n", msg.command)
			command := string(bytes.Trim(msg.command, "\x00"))
			if command == "headers" {
				receivedGetHeader = true
				blocks := ParseGetHeader(msg.payload)
				for i := 0; i < len(blocks); i++ {
					fmt.Printf("block header:\n%s\n", blocks[i])
				}
			}
		}
	}
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
	receivedBuf := make([]byte, 0)
	totalLen := 0
	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		totalLen += n
		receivedBuf = append(receivedBuf, buf...)
		if n < 4096 {
			break
		}
	}

	/*
		the peer node may return version and verack
		at once
	*/
	var msgs []*NetworkEnvelope
	parsedLen := 0
	for {
		if parsedLen >= totalLen {
			break
		}
		msg := ParseNetwork(receivedBuf, s.testnet)
		msgs = append(msgs, msg)
	
		if parsedLen < totalLen {
			parsedLen += len(msg.Serialize())
			receivedBuf = receivedBuf[len(msg.Serialize()):]
		}
	}
	return msgs
}
```
In the aboved code, in the Run method of SimpleNode, it calls WaitFor method to initialize the handshaking process, then it calls GetHeaders
to send the getheaders command packet to the given full node peer. In GetHeaders, we ignore any return packets that their command is not
"header", if the command of packet is "headers", then the full node peer return block header info as we require.

Notices we make some change to the Read method. Since the size of returned packet may quit large, in order to completely received all
data, we first create a 4k buffer, if the returned data volumn is larger than 4k, then we create another 4k buffer to get the remaining
data. We keep create new 4k buffer as long as there are more than 4k bytes of data we need to read.

Then we use ParseGetHeader method to parse the return payload, We have already discuss about the logic of ParseGetHeader before, when 
running the aboved code, I can get the following result:
```go
header count:2000
block header:
version:00000001
previous block id:000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
merkle root:0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098
time stamp:4966bc61
bits:ffff001d
nonce:01e36299
hash:00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048

....
```
Since my full node peer has already download lots of block data from the chain therefore it can return 2000 block headers to me at once!

