package main

import (
	"log"
)

// autoSetupKeyFromEnv — reads HORIZON_ISSUER_KEY from env
// and sets up the key vault on startup (idempotent)
func autoSetupKeyFromEnv() {
	log.Printf("[KEY-ENV] autoSetupKeyFromEnv is DISABLED. Key must be loaded from USB.")
}
