package engine

type SignatureEngine interface {
	Name() string
	Sign(payload []byte, privateKey []byte) (signature []byte, pubKey []byte, err error)
	Verify(payload []byte, signature []byte, publicKey []byte) (bool, error)
}
