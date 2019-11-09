package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	math_rand "math/rand"

	"goimport.moetang.info/nekoq-security/alg/shamir"
)

func main() {

	for i := 0; i < 10000000; i++ {
		key := make([]byte, 32)
		_, _ = rand.Read(key)

		shares, err := shamir.SplitByShamirString(key, 3, 5)
		if err != nil {
			panic(err)
		}

		for k := 0; k < 50; k++ {
			sl := pick(shares, 3)
			recoveredKey, err := shamir.CombineShamirString(sl)
			if err != nil {
				panic(err)
			}
			if !equalByteArray(key, recoveredKey) {
				fmt.Println(key)
				fmt.Println(shares)
				fmt.Println(recoveredKey)
				panic(errors.New("key not matched"))
			}
		}

		if i%1000 == 999 {
			fmt.Println("run:", i)
		}
	}
}

func pick(sl []string, min int) []string {
	n := math_rand.Intn(len(sl)-min+1) + min

	m := make(map[int]struct{})
	for i := 0; i < n; i++ {
		for {
			idx := math_rand.Intn(len(sl))
			if _, ok := m[idx]; !ok {
				m[idx] = struct{}{}
				break
			}
		}
	}

	r := make([]string, n)
	i := 0
	for k, _ := range m {
		r[i] = sl[k]
		i++
	}
	return sl
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
