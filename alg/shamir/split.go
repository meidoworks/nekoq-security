package shamir

import (
	"errors"
	"math/big"
)

/*
 * Key size must equal to or smaller than 256bit
 * Minimum must equal to or greater than 2 and equal to or smaller than 127
 * Shares must equal to or greater than Minimum and equal to or smaller than 127
 */
func SplitByShamir(key []byte, minimum, shares int) ([][]byte, error) {
	if len(key) > 32 {
		return nil, errors.New("key length should equal to or smaller than 256bit")
	}
	if len(key) <= 0 {
		return nil, errors.New("key length should not zero")
	}
	if minimum < 2 || minimum > 127 {
		return nil, errors.New("minimum is out of bound")
	}
	if shares < minimum || shares > 127 {
		return nil, errors.New("shares is out of bound")
	}

	polyList := make([]*big.Int, minimum)
	for i := 1; i < minimum; i++ {
		polyList[i], _ = randFunc(_13th_Mersenne_Prime_Plus_One)
	}
	var s secret
	s.key = key
	polyList[0] = s.generateInt()

	points := make([]struct {
		Idx int
		Val *big.Int
	}, shares)
	for i := 1; i <= shares; i++ {
		points[i-1] = struct {
			Idx int
			Val *big.Int
		}{
			i,
			evalAt(polyList, i, _13th_Mersenne_Prime),
		}
	}

	r := make([][]byte, len(points))
	for i, v := range points {
		var shareObj shareObj
		shareObj.idx = byte(v.Idx)
		shareObj.share = v.Val.Bytes()
		shareObj.neg = v.Val.Sign() < 0
		r[i] = shareObj.generateOutput()
	}

	return r, nil
}

func evalAt(polyList []*big.Int, i int, prime *big.Int) *big.Int {
	var accum = big.NewInt(0)
	for x := len(polyList) - 1; x >= 0; x-- {
		accum = accum.Mul(accum, big.NewInt(int64(i)))
		accum = accum.Add(accum, polyList[x])
		accum = accum.Mod(accum, prime)
	}
	return accum
}
