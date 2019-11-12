package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"goimport.moetang.info/nekoq-security/alg/shamir"
	"goimport.moetang.info/nekoq-security/api"
)

const _testText string = "This is a test text!"

var (
	_ api.MasterKeyProvider            = new(shamirMasterKeyProvider)
	_ api.MasterKeyProviderInitializer = new(shamirMasterKeyProvider)
)

type shamirMasterKeyProvider struct {
	key    []byte
	shares []string

	aesCipher cipher.Block
}

func NewShamirMasterKeyProvider(minimum, shareNum int) (*shamirMasterKeyProvider, error) {
	// generate AES-256 key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	shares, err := shamir.SplitByShamirString(key, 3, 5)
	if err != nil {
		return nil, err
	}

	p := new(shamirMasterKeyProvider)
	p.key = key
	p.shares = shares

	return p, nil
}

func OpenShamirMasterKeyProvider(shares []string, testText string) (*shamirMasterKeyProvider, error) {
	recoveredKey, err := shamir.CombineShamirString(shares)
	if err != nil {
		return nil, err
	}

	p := new(shamirMasterKeyProvider)
	p.key = recoveredKey

	d, err := p.Decrypt([]byte(testText))
	if err != nil {
		return nil, err
	}

	// try to validate master key
	if string(d) != _testText {
		return nil, errors.New("master key not valid")
	}

	return p, nil
}

func (this *shamirMasterKeyProvider) GetTestText() string {
	b, err := this.Encrypt([]byte(_testText))
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (this *shamirMasterKeyProvider) GenerateInitializingKey(p interface{}) (interface{}, error) {
	return this.shares, nil
}

func (this *shamirMasterKeyProvider) GetProviderInfo() api.MasterKeyProviderInfo {
	return api.MasterKeyProviderInfo{
		ProviderName: "ShamirMasterKeyProvider",
		Active:       len(this.key) > 0,
	}
}

func (this *shamirMasterKeyProvider) ensureAes() error {
	if this.aesCipher == nil && len(this.key) > 0 {
		aesCipher, err := aes.NewCipher(this.key)
		if err != nil {
			return err
		}
		this.aesCipher = aesCipher
		return nil
	}

	if this.aesCipher != nil {
		return nil
	}

	if len(this.key) <= 0 {
		return errors.New("cannot init ShamirMasterKey provider")
	}

	panic("should not reach here.")
}

func (this *shamirMasterKeyProvider) Encrypt(dataKey []byte) ([]byte, error) {
	err := this.ensureAes()
	if err != nil {
		return nil, err
	}

	r := make([]byte, len(dataKey))
	this.aesCipher.Encrypt(r, dataKey)

	return r, nil
}

func (this *shamirMasterKeyProvider) Decrypt(encryptedText []byte) ([]byte, error) {
	err := this.ensureAes()
	if err != nil {
		return nil, err
	}

	r := make([]byte, len(encryptedText))
	this.aesCipher.Decrypt(r, encryptedText)

	return r, nil
}
