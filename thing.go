package main

import "fmt"
import "crypto/sha256"
import "math/rand"
import "encoding/binary"
import "bytes"
//import "time"
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
	nonce := make([]byte,64)
	binary.PutUvarint(nonce,b.nonce)
	difficultyAddition := make([]byte,32)
	binary.PutUvarint(difficultyAddition,b.difficultyAddition)
	return sha256.Sum256(append(append(append(b.data,b.prevHash[:]...),nonce...),difficultyAddition...))
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
	data := make([]byte,64)
	conn.Read(data)
	if data[0] == 48 { //send length of chain
		fmt.Fprint(conn,len(blockchain))
		conn.Close()
	}
	fmt.Println(data)
}

func main() {
	blockchain=[]block{block{[]byte{},[hashSize]byte{},0,0}}
	addToChain([]byte("a"))
	addToChain([]byte("b"))
	addToChain([]byte("c"))

	nodelist = os.Args[1:]
	//var lengths []int
	for _,node := range nodelist {
		conn,err := net.Dial("tcp",node)
		if err == nil {
			//fmt.Fprint(conn,'\x01')
			binary.Write(conn,binary.LittleEndian,65)
			conn.Close()
			//buf := make([]byte,32)
			//conn.Read(buf)
			//fmt.Println(buf)
		}
	}

	listen,_ := net.Listen("tcp", ":6565")
	for {
		conn,_ := listen.Accept()
		go handleConn(conn)
	}
}
