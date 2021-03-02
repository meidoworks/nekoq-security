package shamir

import "crypto/rand"

func InitShamirKeys(max, min int) ([]string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	shares, err := SplitByShamirString(key, min, max)
	if err != nil {
		return nil, err
	}

	return shares, nil
}
