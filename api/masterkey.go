package api

type MasterKeyProviderInfo struct {
	ProviderName string
	Active       bool // whether active or not
}

type MasterKeyProvider interface {
	GetProviderInfo() MasterKeyProviderInfo
	GetTestText() string

	Encrypt(dataKey []byte) ([]byte, error)
	Decrypt(encryptedText []byte) ([]byte, error)
}

type MasterKeyProviderInitializer interface {
	GenerateInitializingKey(p interface{}) (interface{}, error)
}
