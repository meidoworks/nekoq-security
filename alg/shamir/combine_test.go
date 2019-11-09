package shamir

import "testing"

func TestCombineShamir(t *testing.T) {
	vl, err := SplitByShamir([]byte{5, 6, 7, 8}, 3, 5)
	if err != nil {
		t.Fatal(err)
	}

	key, err := CombineShamir(vl)
	if err != nil {
		t.Fatal(err)
	}

	if !equalByteArray(key, []byte{5, 6, 7, 8}) {
		t.Fatal("recovered key not matched")
	}
}
