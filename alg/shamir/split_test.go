package shamir

import (
	"encoding/base64"
	"testing"
)

func TestSplitByShamir(t *testing.T) {
	vl, err := SplitByShamir([]byte{5, 6, 7, 8}, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vl {
		t.Log(base64.StdEncoding.EncodeToString(v))
	}
}
