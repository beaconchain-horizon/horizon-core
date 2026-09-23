package license

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type ProofStep struct {
	Sibling string `json:"sibling"`
	Left    bool   `json:"left"`
}

type PrepaidPackage struct {
	Volume int      `json:"volume"`
	Seed   []byte   `json:"seed"`
	Leaves [][]byte `json:"leaves"`
	Root   []byte   `json:"root"`
}

func NewPrepaidPackage(volume int) *PrepaidPackage {
	p := &PrepaidPackage{Volume: volume}
	if volume > 0 {
		p.Leaves = make([][]byte, 0, volume)
	}
	return p
}

func (p *PrepaidPackage) Generate() error {
	if p.Volume <= 0 {
		return errors.New("volume must be greater than zero")
	}
	p.Seed = make([]byte, 32)
	if _, err := rand.Read(p.Seed); err != nil {
		return err
	}
	p.Leaves = make([][]byte, p.Volume)
	for i := range p.Leaves {
		h := sha256.Sum256(append(append([]byte{}, p.Seed...), byte(i>>24), byte(i>>16), byte(i>>8), byte(i)))
		p.Leaves[i] = h[:]
	}
	level := append([][]byte(nil), p.Leaves...)
	for len(level) > 1 {
		next := make([][]byte, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := left
			if i+1 < len(level) {
				right = level[i+1]
			}
			h := sha256.Sum256(append(append([]byte{}, left...), right...))
			next = append(next, h[:])
		}
		level = next
	}
	p.Root = append([]byte(nil), level[0]...)
	return nil
}

func (p *PrepaidPackage) RootHex() string { return hex.EncodeToString(p.Root) }
func (p *PrepaidPackage) SeedHex() string { return hex.EncodeToString(p.Seed) }

func HashHex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
