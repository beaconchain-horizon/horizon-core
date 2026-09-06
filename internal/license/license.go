package license

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type PrepaidPackage struct {
	Volume int
	Seed   []byte
	Root   []byte
}

func NewPrepaidPackage(volume int) *PrepaidPackage {
	if volume <= 0 {
		volume = 1
	}
	seed := make([]byte, 32)
	hash := sha256.Sum256([]byte("HorizonSeed" + time.Now().String()))
	copy(seed, hash[:])
	return &PrepaidPackage{Volume: volume, Seed: seed}
}

func (p *PrepaidPackage) Generate() error {
	leaves := make([][]byte, p.Volume)
	for i := 0; i < p.Volume; i++ {
		leafData := append(p.Seed, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
		hash := sha256.Sum256(leafData)
		leaves[i] = hash[:]
	}
	level := leaves
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := append(level[i], level[i+1]...)
				hash := sha256.Sum256(combined)
				next = append(next, hash[:])
			} else {
				next = append(next, level[i])
			}
		}
		level = next
	}
	p.Root = level[0]
	return nil
}

func (p *PrepaidPackage) RootHex() string {
	return hex.EncodeToString(p.Root)
}

func (p *PrepaidPackage) SeedHex() string {
	return hex.EncodeToString(p.Seed)
}

type LicenseInfo struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
	Signature string    `json:"signature"`
	RootHash  string    `json:"root_hash"`
}

func (l *LicenseInfo) ToJSON() []byte {
	data, _ := json.Marshal(l)
	return data
}
