package main

import (
	"networking"
)

func main() {
	// networkRawData, err := hex.DecodeString("f9beb4d976657273696f6e0000000000650000005f1a69d2721101000100000000000000bc8f5e5400000000010000000000000000000000000000000000ffffc61b6409208d010000000000000000000000000000000000ffffcb0071c0208d128035cbc97953f80f2f5361746f7368693a302e392e332fcf05050001")
	// if err != nil {
	// 	panic(err)
	// }

	// network := networking.ParseNetwork(networkRawData, false)
	// fmt.Printf("%s\n", network)
	// fmt.Printf("%x\n", network.Serialize())
	// version := networking.NewVersionMessage()
	// fmt.Printf("version: %x\n", version.Serialize())

	simpleNode := networking.NewSimpleNode("192.168.3.6", 8333, false)
	simpleNode.Run()

	// getHeaderMsg := networking.NewGetHeaderMessage(networking.GetGenesisBlockHash())
	// fmt.Printf("raw data for get headers msg:%x\n", getHeaderMsg.Serialize())
}
