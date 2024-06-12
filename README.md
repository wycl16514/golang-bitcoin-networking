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

```






