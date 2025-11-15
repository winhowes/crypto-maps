package main

import (
	"fmt"
	"log"

	"crypto-maps-playground/go/circuit"
	"crypto-maps-playground/go/ges"
	"crypto-maps-playground/go/we"
)

func main() {
	// 1. Initialize graded encoding context (CLT13 via ges).
	// These params are placeholders; tune them to match your CLT13 wrapper.
	params := ges.Params{
		Lambda: 128,
		Kappa:  3,   // enough for a tiny AND circuit
		NZs:    1,   // simplest index vector
		Pows:   []uint32{3}, // top-level index; adjust for your CLT13 config
		Flags:  0,
	}
	ctx, err := ges.NewContext(params)
	if err != nil {
		log.Fatalf("NewContext: %v", err)
	}
	defer ctx.Free()

	// 2. Build a simple (a AND b) circuit.
	circ := circuit.ExampleAndCircuit()
	stmt := &we.Statement{Circuit: circ}

	msg := []byte("secret message gated by (a AND b)")

	// 3. Encrypt under the statement.
	ct, err := we.Encrypt(stmt, ctx, msg)
	if err != nil {
		log.Fatalf("we.Encrypt: %v", err)
	}
	fmt.Println("Ciphertext created.")

	// 4. Try decrypt with various witnesses.
	tests := [][]byte{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	for _, w := range tests {
		dec, err := we.Decrypt(stmt, ctx, w, ct)
		if err != nil {
			fmt.Printf("witness %v: decrypt error: %v\n", w, err)
			continue
		}
		fmt.Printf("witness %v: got plaintext %q\n", w, string(dec))
	}
}
