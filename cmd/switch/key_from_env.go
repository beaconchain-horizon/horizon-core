package main

import (
	"log"
	"os"
	"time"

)

// autoSetupKeyFromEnv — reads HORIZON_ISSUER_KEY from env
// and sets up the key vault on startup (idempotent)
func autoSetupKeyFromEnv() {
	hexKey := os.Getenv("HORIZON_ISSUER_KEY")
	if hexKey == "" {
		log.Printf("[KEY-ENV] HORIZON_ISSUER_KEY not set")
		return
	}
	if len(hexKey) < 64 {
		log.Printf("[KEY-ENV] HORIZON_ISSUER_KEY too short: %d", len(hexKey))
		return
	}

	// check if a key already exists
	var vault KeyVault
	if err := db.First(&vault).Error; err == nil {
		log.Printf("[KEY-ENV] key already set up (addr=%s)", vault.PublicAddr)
		return
	}

	// setup with env key
	priv, err := hexToECDSA(hexKey)
	if err != nil {
		log.Printf("[KEY-ENV] invalid key: %v", err)
		return
	}

	password := os.Getenv("KEY_VAULT_PASSWORD")
	if password == "" {
		password = "testpass123"
	}

	encrypted, iv, err := encryptPrivateKey(hexKey, password)
	if err != nil {
		log.Printf("[KEY-ENV] encrypt failed: %v", err)
		return
	}

	addr := publicAddressFromKey(priv)
	newVault := &KeyVault{
		EncryptedKey: encrypted,
		IV:           iv,
		PublicAddr:   addr,
		Label:        "auto-from-env",
	}
	if err := db.Create(newVault).Error; err != nil {
		log.Printf("[KEY-ENV] save failed: %v", err)
		return
	}

	log.Printf("[KEY-ENV] ✅ key set up from env, addr=%s", addr)

	// auto-unlock
	stateMutex.Lock()
	unlockedKey = priv
	unlockedAt = time.Now()
	unlockedAddr = addr
	stateMutex.Unlock()

	log.Printf("[KEY-ENV] ✅ key unlocked, addr=%s", addr)

	// ensure genesis
	if err := ensureChainGenesis(db, priv); err != nil {
		log.Printf("[KEY-ENV] genesis: %v", err)
	}
}
