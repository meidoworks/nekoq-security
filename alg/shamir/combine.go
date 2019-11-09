package shamir

import (
	"errors"
	"math/big"
)

func CombineShamir(sl [][]byte) ([]byte, error) {
	if len(sl) < 2 {
		return nil, errors.New("input array length must equal or greater than 2")
	}
	for _, v := range sl {
		if len(v) <= 2 {
			return nil, errors.New("one or more share length should greater than 2")
		}
	}

	bl := make([]struct {
		Idx int
		Val *big.Int
	}, len(sl))

	for i, v := range sl {
		var shareObj shareObj
		shareObj.fromInput(v)
		iv := new(big.Int)
		if shareObj.neg {
			iv.Neg(new(big.Int).SetBytes(shareObj.share))
		} else {
			iv.SetBytes(shareObj.share)
		}
		bl[i] = struct {
			Idx int
			Val *big.Int
		}{Idx: int(shareObj.idx) & 0xFF, Val: iv}
	}

	b := recoverSecret(bl)

	var s secret
	s.fromInt(b)
	return s.key, nil
}

func recoverSecret(bytes []struct {
	Idx int
	Val *big.Int
}) *big.Int {
	idxList := make([]int, len(bytes))
	shareList := make([]*big.Int, len(bytes))
	for i, v := range bytes {
		idxList[i] = v.Idx
		shareList[i] = v.Val
	}
	return _lagrange_interpolate(0, idxList, shareList, _13th_Mersenne_Prime)
}

func _lagrange_interpolate(x int, idxList []int, shareList []*big.Int, prime *big.Int) *big.Int {
	l := len(idxList)

	var nums []*big.Int
	var dens []*big.Int
	for i := 0; i < l; i++ {
		nums = append(nums, productOfInput(calDiffArr(x, i, idxList)))
		dens = append(dens, productOfInput(calDiffArr(idxList[i], i, idxList)))
	}
	den := productOfInput(dens)

	var arr []*big.Int
	for i := 0; i < l; i++ {
		arr = append(arr, _divmod(mod(mul(mul(nums[i], den), shareList[i]), prime), dens[i], prime))
	}
	num := sum(arr)

	return mod(plus(_divmod(num, den, prime), prime), prime)
}

func sum(arr []*big.Int) *big.Int {
	r := big.NewInt(0)
	for _, v := range arr {
		r.Add(r, v)
	}
	return r
}

func _divmod(num, den, p *big.Int) *big.Int {
	inv, _ := _extended_gcd(den, p)
	return mul(num, inv)
}

func _extended_gcd(a, b *big.Int) (*big.Int, *big.Int) {
	zero := big.NewInt(0)
	x := big.NewInt(0)
	lastX := big.NewInt(1)
	y := big.NewInt(1)
	lastY := big.NewInt(0)
	for b.Cmp(zero) != 0 {
		quot := div(a, b)
		a, b = b, mod(a, b)
		x, lastX = sub(lastX, mul(quot, x)), x
		y, lastY = sub(lastY, mul(quot, y)), y
	}
	return lastX, lastY
}

func calDiffArr(s, notIdx int, idxList []int) []*big.Int {
	var result []*big.Int
	ss := int64(s)
	for i := 0; i < len(idxList); i++ {
		if i == notIdx {
			continue
		}
		ii := int64(idxList[i])
		result = append(result, big.NewInt(ss-ii))
	}
	return result
}

func productOfInput(inputs []*big.Int) *big.Int {
	accum := big.NewInt(1)
	for _, v := range inputs {
		accum = mul(accum, v)
	}
	return accum
}
