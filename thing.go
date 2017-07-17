package main

import "fmt"
import "crypto/sha256"
import "math/rand"
import "encoding/binary"
import "bytes"
import "time"
/*import "math/big"
import "crypto/elliptic"
import "crypto/ecdsa"
import "io/ioutil"
import "os"*/

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
func bruteforce(data []byte,quitChannel chan bool) (bool,block) {
	currentBlock := block{data,blockchain[len(blockchain)-1].getHash(),rand.Uint64(),0}
	for ;!(currentBlock.checkNonce(difficulty1,difficulty2));currentBlock.nonce++ {
		/*if <-quitChannel {
			return false,currentBlock
		}*/
	}
	return true,currentBlock
}

func main() {
	blockchain=[]block{block{[]byte{},[hashSize]byte{},0,0}}
	addToChain([]byte("a"))
	addToChain([]byte("b"))
	addToChain([]byte("c"))
	quitChannel := make(chan bool,1)
	//go bruteforce([]byte("d"),quitChannel)
	//for {
	//	quitChannel<-false
	//}
	now := time.Now()
	bruteforce([]byte("d"),quitChannel)
	fmt.Println(time.Since(now))
}
