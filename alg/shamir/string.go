package shamir

import "encoding/base64"

func SplitByShamirString(key []byte, minimum, shares int) ([]string, error) {
	vl, err := SplitByShamir(key, minimum, shares)
	if err != nil {
		return nil, err
	}
	r := make([]string, len(vl))
	for i, v := range vl {
		r[i] = base64.StdEncoding.EncodeToString(v)
	}
	return r, nil
}

func CombineShamirString(sl []string) ([]byte, error) {
	d := make([][]byte, len(sl))
	for i, v := range sl {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		d[i] = b
	}

	return CombineShamir(d)
}
