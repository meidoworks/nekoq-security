package shamir

import (
	"testing"
)

func TestSplitByShamirString(t *testing.T) {
	sl, err := SplitByShamirString([]byte{1, 2, 3, 4, 5, 6}, 3, 5)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range sl {
		t.Log(v)
	}
}

func TestCombineShamirString(t *testing.T) {
	sl, err := SplitByShamirString([]byte{1, 2, 3, 4, 5, 6}, 3, 5)
	if err != nil {
		t.Fatal(err)
	}

	key, err := CombineShamirString(sl)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(key)
}

func BenchmarkCombineShamirString(b *testing.B) {
	sl, err := SplitByShamirString([]byte{1, 2, 3, 4, 5, 6}, 3, 5)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := CombineShamirString(sl)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSplitByShamirString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := SplitByShamirString([]byte{1, 2, 3, 4, 5, 6}, 3, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}
