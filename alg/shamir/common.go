package shamir

import (
	"crypto/rand"
	"math/big"
)

const (
	_VERSION byte = 1 // two bits
)

var _13th_Mersenne_Prime *big.Int
var _13th_Mersenne_Prime_Plus_One *big.Int

func init() {
	_13th_Mersenne_Prime = mersenneNumber(521)
	_13th_Mersenne_Prime_Plus_One = new(big.Int).Add(_13th_Mersenne_Prime, big.NewInt(1))
}

func mersenneNumber(m int) *big.Int {
	return new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(m)), big.NewInt(0)), big.NewInt(1))
}

var randFunc = func(i *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, i)
}

func plus(a, b *big.Int) *big.Int {
	acp := new(big.Int).Set(a)
	return acp.Add(acp, b)
}

func sub(a, b *big.Int) *big.Int {
	acp := new(big.Int).Set(a)
	return acp.Sub(acp, b)
}

func mul(a, b *big.Int) *big.Int {
	acp := new(big.Int).Set(a)
	return acp.Mul(acp, b)
}

func div(a, b *big.Int) *big.Int {
	acp := new(big.Int).Set(a)
	return acp.Div(acp, b)
}

func mod(a, b *big.Int) *big.Int {
	acp := new(big.Int).Set(a)
	return acp.Mod(acp, b)
}

type secret struct {
	key     []byte
	version byte // no need to fill this field when generating int
}

// 1 byte header
//     2 bits versoin
//     6 bits zero padding
func (this secret) generateInt() *big.Int {
	this.version = _VERSION
	l := len(this.key)
	data := append([]byte{0}, this.key...)
	if 64-l > 0 { //max 512bit
		data = append(data, make([]byte, 64-l)...)
	}
	data[0] = (this.version << 6) | byte((64-l)&63)
	return new(big.Int).SetBytes(data)
}

func (this *secret) fromInt(i *big.Int) {
	data := i.Bytes()
	version := (data[0] & 192) >> 6
	pl := data[0] & 63
	key := make([]byte, 64-pl)
	copy(key, data[1:])

	this.key = key
	this.version = version
}

// 1 byte header
//     2 bits versoin
//     1 bit neg
type shareObj struct {
	share   []byte
	version byte // no need to fill this field when generating output
	neg     bool
	idx     byte // 1 - 127
}

func (this shareObj) generateOutput() []byte {
	var b byte = 0
	b = b | (this.version << 6)
	if this.neg {
		b = b | 32
	}
	r := make([]byte, len(this.share)+2)
	r[0] = b
	r[1] = this.idx
	copy(r[2:], this.share)
	return r
}

func (this *shareObj) fromInput(data []byte) {
	b := data[0]
	this.version = b >> 6
	this.neg = (b & 32) > 0
	this.idx = data[1]
	r := make([]byte, len(data)-2)
	copy(r, data[2:])
	this.share = r
}
