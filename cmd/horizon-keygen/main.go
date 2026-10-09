package main

import (
	"fmt"
	"horizon-core/internal/crypto"
)

func main() {
	pem, err := crypto.GenerateKeyPair()
	if err != nil {
		panic(err)
	}
	fmt.Print(pem)
}
