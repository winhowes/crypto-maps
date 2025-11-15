package we

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"crypto-maps-playground/go/circuit"
	"crypto-maps-playground/go/ges"
)

type Statement struct {
	Circuit *circuit.Circuit
	// later: add statement-specific public inputs (e.g. H, genesis hash)
}

type Ciphertext struct {
	// Encodings for input wires, gate gadgets, etc.
	// For now, keep it simple and only support very small circuits.
	WireEncodings [][]*ges.Element // [wire][bitVal] (like wire i, bit 0 or 1)
	GateAux       interface{}      // fill in as needed
	MaskedKey     []byte           // K ⊕ H(g^s)
	SymCiphertext []byte
}

// Encrypt m under "there exists witness w s.t C(x,w)=1"
func Encrypt(stmt *Statement, ctx *ges.Context, msg []byte) (*Ciphertext, error) {
	if stmt == nil || stmt.Circuit == nil {
		return nil, errors.New("statement circuit is required")
	}
	_ = ctx // placeholder for future graded-encoding usage

	K := make([]byte, 32)
	if _, err := rand.Read(K); err != nil {
		return nil, fmt.Errorf("generate symmetric key: %w", err)
	}

	symCt, err := symEncrypt(K, msg)
	if err != nil {
		return nil, err
	}

	return &Ciphertext{
		MaskedKey:     K,
		SymCiphertext: symCt,
	}, nil
}

// Decrypt with witness bits (e.g. []byte{1,1} for (a,b))
func Decrypt(stmt *Statement, ctx *ges.Context, witness []byte, ct *Ciphertext) ([]byte, error) {
	if stmt == nil || stmt.Circuit == nil {
		return nil, errors.New("statement circuit is required")
	}
	if ct == nil {
		return nil, errors.New("ciphertext is required")
	}
	_ = ctx // placeholder for future graded-encoding usage

	out, err := circuit.Eval(stmt.Circuit, witness)
	if err != nil {
		return nil, err
	}
	if out == 0 {
		return nil, errors.New("invalid witness: circuit evaluated to 0")
	}

	if len(ct.MaskedKey) == 0 {
		return nil, errors.New("ciphertext missing masked key")
	}

	msg, err := symDecrypt(ct.MaskedKey, ct.SymCiphertext)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func symEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

func symDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ct := ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
