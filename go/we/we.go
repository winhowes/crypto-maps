package we

import (
    "crypto/rand"
    "crypto/sha256"
    "errors"

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
    // 1. Sample secret s (just a random 256-bit value we’ll implicitly embed)
    s := make([]byte, 32)
    if _, err := rand.Read(s); err != nil {
        return nil, err
    }

    // 2. Sample random symmetric key K
    K := make([]byte, 32)
    if _, err := rand.Read(K); err != nil {
        return nil, err
    }

    // 3. Build encodings for inputs / gates using ges.Context.
    // This is where we replicate the “wire exponents” idea in terms of CLT13 encodings.
    // For first pass, maybe only support ExampleAndCircuit and hardcode.
    wireEncodings, gateAux, gToS, err := buildEncodingsForCircuit(stmt.Circuit, ctx, s)
    if err != nil {
        return nil, err
    }

    // 4. Derive mask = H(g^s)
    T := sha256.Sum256(gToS)

    // 5. Mask K
    maskedKey := make([]byte, len(K))
    for i := range K {
        maskedKey[i] = K[i] ^ T[i]
    }

    // 6. Symmetric encrypt msg under K (use your favorite AEAD)
    symCt, err := symEncrypt(K, msg)
    if err != nil {
        return nil, err
    }

    return &Ciphertext{
        WireEncodings: wireEncodings,
        GateAux:       gateAux,
        MaskedKey:     maskedKey,
        SymCiphertext: symCt,
    }, nil
}

// Decrypt with witness bits (e.g. []byte{1,1} for (a,b))
func Decrypt(stmt *Statement, ctx *ges.Context, witness []byte, ct *Ciphertext) ([]byte, error) {
    if len(witness) != stmt.Circuit.NumInputs {
        return nil, errors.New("witness length mismatch")
    }

    // 1. Evaluate circuit over graded encodings using witness:
    gToS, err := evaluateCircuit(stmt.Circuit, ctx, witness, ct.WireEncodings, ct.GateAux)
    if err != nil {
        return nil, err
    }

    // 2. Derive T' = H(g^s') (or H(1) if unsatisfying)
    T := sha256.Sum256(gToS)

    // 3. Recover K' = maskedKey ⊕ T'
    Kprime := make([]byte, len(ct.MaskedKey))
    for i := range ct.MaskedKey {
        Kprime[i] = ct.MaskedKey[i] ^ T[i]
    }

    // 4. Symmetric decrypt
    msg, err := symDecrypt(Kprime, ct.SymCiphertext)
    if err != nil {
        return nil, err
    }
    return msg, nil
}
