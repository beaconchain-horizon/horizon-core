package engine

import "errors"

type PQCGlobalEngine struct{}
func (p *PQCGlobalEngine) Name() string { return "PQC-Hybrid-Global" }
func (p *PQCGlobalEngine) Sign(payload []byte, privateKey []byte) ([]byte, []byte, error) {
	return nil, nil, errors.New("PQC engine requires external library")
}
func (p *PQCGlobalEngine) Verify(payload []byte, signature []byte, publicKey []byte) (bool, error) {
	return false, errors.New("PQC verify not implemented")
}
