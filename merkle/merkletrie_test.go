package merkle

import (
	"testing"

	"bytes"

	"crypto/rand"

	"math"

	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/thunderdb/ThunderDB/crypto/hash"
)

func TestMergeTwoHash(t *testing.T) {
	Convey("Concatenate of two hash should be equal to the result", t, func() {
		h0 := &hash.Hash{}
		rand.Read(h0[:])
		h1 := &hash.Hash{}
		rand.Read(h1[:])
		wanted := mergeHash(h0, h1)

		So(bytes.Compare(MergeTwoHash(h0, h1).CloneBytes(), wanted[:]), ShouldEqual, 0)
	})
}

func TestNewMerkle(t *testing.T) {
	tests := [][]*hash.Hash{
		[]*hash.Hash{},
		[]*hash.Hash{
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
		},
		[]*hash.Hash{
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
		},
		[]*hash.Hash{
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
			&hash.Hash{},
		},
	}
	Convey("Two root hashes should be the same", t, func() {
		for i := range tests {
			for j := range tests[i] {
				rand.Read(tests[i][j][:])
			}
			merkle := NewMerkle(tests[i])
			wantedMerkle := buildMerkleTreeStore(tests[i])
			wanted := wantedMerkle[len(wantedMerkle)-1]
			root := merkle.GetRoot()
			So(bytes.Compare(wanted[:], root[:]), ShouldEqual, 0)
		}
	})
}

func mergeHash(h0 *hash.Hash, h1 *hash.Hash) *hash.Hash {
	h := hash.THashH(append(h0[:], h1[:]...))
	return &h
}

// nextPowerOfTwo and BuildMerkleTreeStore is copy from github.com/btcsuite/btcd/blockchain/merkle.go
// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

func buildMerkleTreeStore(hashes []*hash.Hash) []*hash.Hash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	if hashes == nil || len(hashes) == 0 {
		hashes = []*hash.Hash{&hash.Hash{}}
	}

	nextPoT := nextPowerOfTwo(len(hashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*hash.Hash, arraySize)

	// Create the base transaction hashes and populate the array with them.
	for i, h := range hashes {
		merkles[i] = h
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

			// When there is no right child, the parent is generated by
			// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := mergeHash(merkles[i], merkles[i])
			merkles[offset] = newHash

			// The normal case sets the parent node to the double sha256
			// of the concatentation of the left and right children.
		default:
			newHash := mergeHash(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}
