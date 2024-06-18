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

```

