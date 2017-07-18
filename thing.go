package main

import "fmt"
import "crypto/sha256"
import "math/rand"
import "encoding/binary"
import "bytes"
import "time"
import "os"
import "net"
/*import "math/big"
import "crypto/elliptic"
import "crypto/ecdsa"
import "io/ioutil"*/

const hashSize = 32

//blocks, chains
type block struct {
	data []byte
	prevHash [hashSize]byte
	nonce uint64
	difficultyAddition uint64
}

func (b block) getHash() [hashSize]byte {
	nonce := make([]byte,8)
	binary.PutUvarint(nonce,b.nonce)
	difficultyAddition := make([]byte,8)
	binary.PutUvarint(difficultyAddition,b.difficultyAddition)
	return sha256.Sum256(append(append(append(b.data,b.prevHash[:]...),nonce...),difficultyAddition...))
}

func bytesToBlock(data []byte) block {
	nonce,_ := binary.ReadUvarint(bytes.NewReader(data[len(data)-8:]))
	return block{data[:len(data)-8],[hashSize]byte{},nonce,uint64(0)}
}
func bytesToChain(data []byte) []block {
	var chain []block
	for ;len(data)>0; {
		len,_ := binary.ReadUvarint(bytes.NewReader(data[:64]))
		data = data[64:]
		chain = append(chain,bytesToBlock(data[:len]))
		data = data[len:]
	}
	return chain
}
func blockToBytes(b block) []byte {
	var buf []byte
	binary.Write(bytes.NewBuffer(buf),binary.LittleEndian,b.nonce)
	return append(b.data,buf...)
}
func chainToBytes(chain []block) []byte {
	var encodedChain []byte
	for _,block := range chain {
		encodedBlock := blockToBytes(block)
		var buf []byte
		binary.Write(bytes.NewBuffer(buf),binary.LittleEndian,uint64(len(encodedBlock)))
		encodedChain = append(encodedChain,buf...)
		encodedChain = append(encodedChain,encodedBlock...)
	}
	return encodedChain
}

var blockchain []block
func addToChain(data []byte) {
	blockchain=append(blockchain,block{data,blockchain[len(blockchain)-1].getHash(),0,0})
}

var difficulty1 uint32 = 0xffefffff
var difficulty2 uint64 = 1
//time to mine block = time to try 1 hash * ((2^32-1)/(2^32-1-difficulty1))^difficulty2
func (b block) checkNonce(d1 uint32,d2 uint64) bool {
	for b.difficultyAddition=0;b.difficultyAddition<d2;b.difficultyAddition++ {
		var hash uint32
		gottenHash := b.getHash()
		binary.Read(bytes.NewReader(gottenHash[:]),binary.LittleEndian,&hash)

		if hash<d1 {
			return false
		}
	}
	return true
}
func bruteforce(data []byte,quitChannel chan struct{}) (bool,block) {
	currentBlock := block{data,blockchain[len(blockchain)-1].getHash(),uint64(rand.Uint32())*uint64(rand.Uint32()),0}
	for {
		select {
			case <-quitChannel:
				return false,currentBlock
			default:
				currentBlock.nonce++
				if currentBlock.checkNonce(difficulty1,difficulty2) {
					return true,currentBlock
				}
		}
	}
}

var nodelist []string
func handleConn(conn net.Conn) {
	data := make([]byte,1)
	conn.Read(data)
	if data[0] == 0 { //send length of chain
		binary.Write(conn,binary.LittleEndian,uint32(len(blockchain)))
		conn.Close()
	}
	fmt.Println(data)
}

func main() {
	//make genesis block
	blockchain=[]block{block{[]byte{},[hashSize]byte{},0,0}}

	nodelist = os.Args[1:]

	//get lengths
	lengths := make([]uint32,len(nodelist))
	max := uint32(0)
	done := 0
	start := time.Now()
	for index,node := range nodelist {
		go func(index int,node string){
			conn,err := net.Dial("tcp",node)
			if err == nil {
				conn.Write([]byte{0})
				var length uint32
				binary.Read(conn,binary.LittleEndian,&length)
				lengths[index]=length
				if length>max {
					max = length
				}
				conn.Close()
				done++
			}
		}(index,node)
	}
	for ;done<len(nodelist) && time.Since(start)<3000000000; {} //wait until 3 seconds pass or all data is gathered
	fmt.Println(lengths)
	//now use counting sort to sort nodelist based on lengths
	histogram := make([]uint32,max+1)
	sortedNodelist := make([]string,len(nodelist))
	for index,_ := range nodelist {
		histogram[lengths[index]]++
	}
	total := uint32(0)
	for index,count := range histogram {
		histogram[index]=total
		total+=count
	}
	for index,node := range nodelist {
		sortedNodelist[histogram[lengths[index]]] = node
		histogram[lengths[index]]++
	}
	fmt.Println(sortedNodelist)

	listen,_ := net.Listen("tcp",":6565")
	for {
		conn,_ := listen.Accept()
		go handleConn(conn)
	}
}
