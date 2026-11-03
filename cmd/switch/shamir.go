package main

import (
	"crypto/rand"
	"fmt"
)

// GF(2^8) arithmetic tables with primitive polynomial 0x11B (x^8 + x^4 + x^3 + x + 1).
// This is the same field used by AES; 3 is a primitive element.
var gfExp [512]byte
var gfLog [256]byte

func init() {
	x := byte(1)
	for i := 0; i < 255; i++ {
		gfExp[i] = x
		gfLog[x] = byte(i)
		x = gfMulSlow(x, 3)
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

// gfMulSlow performs multiplication in GF(2^8) without lookup tables.
func gfMulSlow(a, b byte) byte {
	var p byte
	for i := 0; i < 8; i++ {
		if b&1 != 0 {
			p ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1B
		}
		b >>= 1
	}
	return p
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

func gfDiv(a, b byte) byte {
	if b == 0 {
		panic("GF(2^8) division by zero")
	}
	if a == 0 {
		return 0
	}
	return gfExp[(int(gfLog[a])-int(gfLog[b])+255)%255]
}

// shamirSplit performs real Shamir's Secret Sharing over GF(2^8).
//
// Splits `secret` into `total` shares; any `threshold` of them reconstruct it.
// Security: fewer than `threshold` shares reveal ZERO information about the secret.
//
// Format of each share: [x-coordinate (1..total)] || [y_0, y_1, ..., y_{n-1}]
func shamirSplit(secret []byte, total, threshold int) ([][]byte, error) {
	if threshold > total || threshold < 2 {
		return nil, fmt.Errorf("invalid threshold %d/%d (need 2 <= threshold <= total)", threshold, total)
	}
	if total > 255 {
		return nil, fmt.Errorf("total shares cannot exceed 255 (GF(2^8) limit)")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("secret cannot be empty")
	}

	shares := make([][]byte, total)
	for i := range shares {
		shares[i] = make([]byte, len(secret)+1)
		shares[i][0] = byte(i + 1) // x-coordinate: 1..total (0 reserved for the secret)
	}

	for j, s := range secret {
		coeffs := make([]byte, threshold)
		coeffs[0] = s
		if threshold > 1 {
			if _, err := rand.Read(coeffs[1:]); err != nil {
				return nil, fmt.Errorf("random coefficient generation failed: %w", err)
			}
		}
		// Evaluate polynomial f(x) = c0 + c1*x + ... + c_{t-1}*x^{t-1} at x = i+1
		for i := 0; i < total; i++ {
			x := byte(i + 1)
			y := coeffs[threshold-1]
			for k := threshold - 2; k >= 0; k-- {
				y = gfMul(y, x) ^ coeffs[k]
			}
			shares[i][j+1] = y
		}
	}
	return shares, nil
}

// shamirCombine reconstructs the secret from at least `threshold` shares
// created by shamirSplit. Uses Lagrange interpolation at x=0.
func shamirCombine(shares [][]byte) ([]byte, error) {
	if len(shares) < 2 {
		return nil, fmt.Errorf("at least 2 shares required")
	}
	secretLen := len(shares[0]) - 1
	if secretLen < 0 {
		return nil, fmt.Errorf("share too short")
	}
	for _, s := range shares {
		if len(s) != len(shares[0]) {
			return nil, fmt.Errorf("share length mismatch")
		}
	}

	secret := make([]byte, secretLen)
	for j := 0; j < secretLen; j++ {
		var s byte
		for i, si := range shares {
			xi := si[0]
			yi := si[j+1]

			// Lagrange basis at x=0
			var num, den byte = 1, 1
			for k, sk := range shares {
				if i == k {
					continue
				}
				xk := sk[0]
				num = gfMul(num, xk)
				den = gfMul(den, xi^xk)
			}
			s ^= gfMul(yi, gfDiv(num, den))
		}
		secret[j] = s
	}
	return secret, nil
}

// randomBytes returns n cryptographically secure random bytes.
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
