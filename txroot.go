package main

import (
	"encoding/binary"
	"fmt"

	"github.com/prysmaticlabs/gohashtree"
)

var zero_hash = make([][32]byte, 30)
var test_txs = make([][]byte, 10)

func init() {
	for i := 1; i < len(zero_hash); i++ {
		chunk := make([]byte, 64)
		copy(chunk, zero_hash[i][:])
		copy(chunk[32:], zero_hash[i][:])
		gohashtree.Hash(zero_hash[i][:], chunk)
	}

	for i := 0; i < len(test_txs); i++ {
		test_txs[i] = make([]byte, 64)
		test_txs[i][0] = byte(i)
	}
}

// hashes a slice with limit = 2^depth, destroys input if larger than 64 bytes. input is assumed to be
// of multiple of 32 bytes. You're responsible to pad by zeroes.
func ssz_byte_slice_htr(input []byte, depth int) [32]byte {
	length := (len(input) + 31) / 32
	if length&1 == 1 {
		input = append(input, zero_hash[0][:]...)
		length++
	}
	for layer := 1; layer <= depth; layer++ {
		gohashtree.Hash(input, input[:32*length])
		length = length / 2
		if length&1 == 1 {
			copy(input[32*length:32*length+32], zero_hash[layer][:])
			length++
		}
	}
	ret := [32]byte{}
	copy(ret[:], input[:32])
	return ret
}

// input is assumed to be 32 bytes, returns 32 bytes
func mix_in_length(input []byte, length uint64) []byte {
	buf := [32]byte{}
	binary.LittleEndian.PutUint64(buf[:], length)
	input = append(input, buf[:]...)
	gohashtree.Hash(buf[:], input)
	return buf[:]
}

func main() {
	hash_slice := make([]byte, 32*len(test_txs))
	for _, tx := range test_txs {
		length := len(tx)
		hash := ssz_byte_slice_htr(tx, 25)
		hash_slice = append(hash_slice, mix_in_length(hash[:], uint64(length))...)
	}
	hash := ssz_byte_slice_htr(hash_slice, 20)
	htr := mix_in_length(hash[:], uint64(len(test_txs)))
	fmt.Printf("transactions_root: %x#\n", htr)
}
