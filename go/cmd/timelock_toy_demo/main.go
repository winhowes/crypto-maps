package main

import (
	"fmt"
	"log"

	"crypto-maps-playground/go/ges"
	"crypto-maps-playground/go/timelock"
	"crypto-maps-playground/go/we"
)

func main() {
	// 1. Initialize graded encoding context.
	params := ges.Params{
		Lambda: 128,
		Kappa:  3,
		NZs:    1,
		Pows:   []uint32{3},
		Flags:  0,
	}
	ctx, err := ges.NewContext(params)
	if err != nil {
		log.Fatalf("NewContext: %v", err)
	}
	defer ctx.Free()

	// 2. Define toy time-lock parameters.
	genesis := timelock.GenesisFromString("toy-genesis")
	tle := timelock.TLEParams{
		TargetHeight:   3,
		GenesisHash:    genesis,
		DifficultyBits: 12, // small so mining finishes quickly
	}

	// 3. Encrypt a message under the toy time-lock.
	plaintext := []byte("this is time-locked under a toy chain")
	stmt, ct, err := timelock.EncryptToyTimeLock(ctx, tle, plaintext)
	if err != nil {
		log.Fatalf("EncryptToyTimeLock: %v", err)
	}
	fmt.Println("Toy time-lock ciphertext created.")

	// 4. Try decrypt with an invalid/short chain — should fail.
	shortChain, err := timelock.MineToyChain(genesis, 1, tle.DifficultyBits)
	if err != nil {
		log.Fatalf("MineToyChain(short): %v", err)
	}
	_, err = timelock.DecryptToyTimeLock(ctx, tle, shortChain, stmt, ct)
	if err != nil {
		fmt.Printf("As expected, decryption with short chain failed: %v\n", err)
	} else {
		fmt.Println("Unexpected: decryption succeeded with short chain!")
	}

	// 5. Mine a full-valid chain and decrypt.
	fullChain, err := timelock.MineToyChain(genesis, tle.TargetHeight, tle.DifficultyBits)
	if err != nil {
		log.Fatalf("MineToyChain(full): %v", err)
	}
	recovered, err := timelock.DecryptToyTimeLock(ctx, tle, fullChain, stmt, ct)
	if err != nil {
		log.Fatalf("DecryptToyTimeLock(full): %v", err)
	}
	fmt.Printf("Decryption with valid chain succeeded: %q\n", string(recovered))

	// Optional sanity: ensure recovered == plaintext.
	if string(recovered) != string(plaintext) {
		log.Fatalf("mismatch: recovered %q != original %q", recovered, plaintext)
	}
	fmt.Println("Round-trip success.")
	_ = we.Statement{} // just to avoid unused import errors if you tweak things
}
