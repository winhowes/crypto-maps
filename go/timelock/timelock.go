package timelock

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"crypto-maps-playground/go/circuit"
	"crypto-maps-playground/go/ges"
	"crypto-maps-playground/go/we"
)

// TLEParams describes the toy time-lock condition:
// "there exists a valid chain of length >= TargetHeight extending GenesisHash with PoW difficulty DifficultyBits".
type TLEParams struct {
	TargetHeight   int
	GenesisHash    [ToyHashSize]byte
	DifficultyBits uint8
}

// buildTrivialCircuit returns a 1-input circuit that just forwards the bit as output.
// Relation: "witness_bit == 1".
//
// This is a *placeholder*; eventually you’d replace this with a real circuit that
// verifies the toy chain inside the WE relation.
func buildTrivialCircuit() *circuit.Circuit {
	// One input a; one gate G0: INPUT (just passes a); output = G0
	return &circuit.Circuit{
		NumInputs: 1,
		Gates: []circuit.Gate{
			{
				Type: circuit.GateInput,
				In1:  0,
				In2:  -1,
			},
		},
		OutputGate: 0,
	}
}

// EncryptToyTimeLock encrypts msg under the condition "there exists a valid toy chain meeting TLEParams".
//
// Internally, this uses a trivial 1-bit circuit and relies on the caller to only ever
// provide witness bit 1 when VerifyToyChain(...) succeeds. It’s a good stepping stone
// before you encode the full chain relation into a circuit.
func EncryptToyTimeLock(ctx *ges.Context, params TLEParams, msg []byte) (*we.Statement, *we.Ciphertext, error) {
	// For now, the statement is just the circuit. Later, you can add params into the statement.
	circ := buildTrivialCircuit()
	stmt := &we.Statement{Circuit: circ}

	ct, err := we.Encrypt(stmt, ctx, msg)
	if err != nil {
		return nil, nil, fmt.Errorf("we.Encrypt: %w", err)
	}
	return stmt, ct, nil
}

// DecryptToyTimeLock verifies the given chain and, if valid, treats that as a witness
// for the trivial circuit (witness bit = 1). If the chain is invalid or too short,
// decryption fails.
func DecryptToyTimeLock(ctx *ges.Context, params TLEParams, chain *ToyChain, stmt *we.Statement, ct *we.Ciphertext) ([]byte, error) {
	// 1. Verify chain in "native" Go.
	if err := VerifyToyChain(chain, params.GenesisHash, params.DifficultyBits, params.TargetHeight); err != nil {
		return nil, fmt.Errorf("VerifyToyChain: %w", err)
	}

	// 2. Feed witness bit = 1 into the trivial circuit.
	witness := []byte{1}
	msg, err := we.Decrypt(stmt, ctx, witness, ct)
	if err != nil {
		return nil, fmt.Errorf("we.Decrypt: %w", err)
	}
	return msg, nil
}

// GenesisFromString deterministically derives a genesis hash from the provided string.
func GenesisFromString(s string) [ToyHashSize]byte {
	return sha256.Sum256([]byte(s))
}

// RandomGenesis samples a uniformly random genesis hash.
func RandomGenesis() ([ToyHashSize]byte, error) {
	var out [ToyHashSize]byte
	if _, err := rand.Read(out[:]); err != nil {
		return out, fmt.Errorf("rand.Read genesis: %w", err)
	}
	return out, nil
}
