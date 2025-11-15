package timelock

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

const ToyHashSize = 32 // bytes

// ToyBlock is a tiny "block header" for a toy PoW chain.
type ToyBlock struct {
	PrevHash [ToyHashSize]byte
	Nonce    uint64
	Hash     [ToyHashSize]byte // hash of (PrevHash || Nonce)
}

// ToyChain is just a slice of blocks.
type ToyChain struct {
	Blocks []ToyBlock
}

// NewToyChain creates a chain with just the genesis block hash recorded.
func NewToyChain(genesis [ToyHashSize]byte) *ToyChain {
	// We treat genesis as "height 0"; first mined block has PrevHash = genesis.
	return &ToyChain{Blocks: make([]ToyBlock, 0)}
}

// hashHeader computes H(PrevHash || Nonce).
func hashHeader(prev [ToyHashSize]byte, nonce uint64) [ToyHashSize]byte {
	var buf [ToyHashSize + 8]byte
	copy(buf[:ToyHashSize], prev[:])
	binary.BigEndian.PutUint64(buf[ToyHashSize:], nonce)
	return sha256.Sum256(buf[:])
}

// checkDifficulty returns true if the hash has at least difficultyBits leading zero bits.
func checkDifficulty(h [ToyHashSize]byte, difficultyBits uint8) bool {
	fullBytes := int(difficultyBits / 8)
	remBits := int(difficultyBits % 8)

	// Check full zero bytes
	for i := 0; i < fullBytes; i++ {
		if h[i] != 0 {
			return false
		}
	}
	if remBits == 0 {
		return true
	}

	// Check partial byte
	mask := byte(0xFF << (8 - remBits))
	return (h[fullBytes] & mask) == 0
}

// MineNextBlock mines a single block extending the given prevHash.
func MineNextBlock(prev [ToyHashSize]byte, difficultyBits uint8) (ToyBlock, error) {
	var blk ToyBlock
	blk.PrevHash = prev

	// Start nonce from random point to vary the search.
	var startNonce uint64
	if err := binary.Read(rand.Reader, binary.BigEndian, &startNonce); err != nil {
		return ToyBlock{}, fmt.Errorf("rand nonce: %w", err)
	}

	nonce := startNonce
	for {
		h := hashHeader(prev, nonce)
		if checkDifficulty(h, difficultyBits) {
			blk.Nonce = nonce
			blk.Hash = h
			return blk, nil
		}
		nonce++
	}
}

// MineToyChain mines a chain of length targetHeight over the given genesis hash.
// DifficultyBits is intentionally small (e.g., 8–16) so this finishes in seconds.
func MineToyChain(genesis [ToyHashSize]byte, targetHeight int, difficultyBits uint8) (*ToyChain, error) {
	if targetHeight <= 0 {
		return nil, errors.New("targetHeight must be > 0")
	}
	chain := &ToyChain{Blocks: make([]ToyBlock, 0, targetHeight)}

	prev := genesis
	for i := 0; i < targetHeight; i++ {
		blk, err := MineNextBlock(prev, difficultyBits)
		if err != nil {
			return nil, err
		}
		chain.Blocks = append(chain.Blocks, blk)
		prev = blk.Hash
	}

	return chain, nil
}

// VerifyToyChain checks:
//   - length >= minHeight
//   - first block's PrevHash = genesis
//   - hash chaining is correct
//   - each block meets difficulty
func VerifyToyChain(chain *ToyChain, genesis [ToyHashSize]byte, difficultyBits uint8, minHeight int) error {
	if chain == nil {
		return errors.New("nil chain")
	}
	if len(chain.Blocks) < minHeight {
		return fmt.Errorf("chain too short: have %d, need at least %d", len(chain.Blocks), minHeight)
	}
	if len(chain.Blocks) == 0 {
		return errors.New("empty chain")
	}

	// First block must extend genesis
	if chain.Blocks[0].PrevHash != genesis {
		return errors.New("first block prev hash != genesis")
	}

	// Check PoW + chaining
	prev := genesis
	for i, blk := range chain.Blocks {
		if blk.PrevHash != prev {
			return fmt.Errorf("block %d prev hash mismatch", i)
		}
		expectedHash := hashHeader(blk.PrevHash, blk.Nonce)
		if blk.Hash != expectedHash {
			return fmt.Errorf("block %d hash mismatch", i)
		}
		if !checkDifficulty(blk.Hash, difficultyBits) {
			return fmt.Errorf("block %d fails difficulty", i)
		}
		prev = blk.Hash
	}
	return nil
}
