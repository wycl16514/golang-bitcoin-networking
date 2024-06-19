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
func ParseGetHeader(rawData []byte) []*tx.Block {
	reader := bytes.NewReader(rawData)
	bufReader := bufio.NewReader(reader)
	numHeaders := tx.ReadVarint(bufReader)
	fmt.Printf("header count:%d\n", numHeaders.Int64())
	//shift to the beginning of block header
	rawData = rawData[len(numHeaders.Bytes()):]
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

		shift := len(numTxs.Bytes())
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

