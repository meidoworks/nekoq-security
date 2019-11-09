package shamir

import (
	"math/big"
	"testing"
)

var t1 *big.Int

func init() {
	var s secret
	s.version = _VERSION
	s.key = []byte{5, 6, 7, 8}

	t1 = s.generateInt()
}

func TestSecretGenerate(t *testing.T) {
	var s secret
	s.version = _VERSION
	s.key = []byte{5, 6, 7, 8}

	if s.generateInt().Cmp(_13th_Mersenne_Prime_Plus_One) >= 0 {
		t.Fatal("generateInt s should less than 13th mersenne prime")
	}

	if !equalByteArray(s.generateInt().Bytes(), []byte{124, 5, 6, 7, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatal("generateInt s output not desired. s output:", s.generateInt().Bytes())
	}
}

func TestSecretFromInt(t *testing.T) {
	var s secret
	s.fromInt(t1)

	if s.version != _VERSION {
		t.Fatal("version s not match _VERSION")
	}
	if !equalByteArray(s.key, []byte{5, 6, 7, 8}) {
		t.Fatal("key not matched. s.key:", s.key)
	}
}

func equalByteArray(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}
