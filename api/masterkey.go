package api

type MasterKeyProviderInfo struct {
	ProviderName string
	Active       bool // whether active or not
}

type MasterKeyProvider interface {
	GetProviderInfo() MasterKeyProvider
	Encrypt(dataKey []byte) ([]byte, error)
	Decrypt(encryptedText []byte) ([]byte, error)
}
