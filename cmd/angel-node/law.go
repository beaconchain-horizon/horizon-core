package main

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "log"
    "strings"
)

const LawVersion = "1.0"

type LawRule struct {
    Role, Port, Equation, Level string
}

var CanonicalLaw = []LawRule{
    {"merkle", "9001", "SHA256(left||right)", "CRITICAL"},
    {"hash", "9002", "SHA256(header)", "CRITICAL"},
    {"ecdsa", "9003", "secp256k1.Verify", "CRITICAL"},
    {"type", "9004", "type in binary or ternary", "CRITICAL"},
    {"volume", "9005", "used < volume", "IMPORTANT"},
    {"time", "9006", "issued <= now < expires", "IMPORTANT"},
    {"user", "9007", "user_id match", "IMPORTANT"},
}

func LawHash() string {
    var sb strings.Builder
    sb.WriteString("v=" + LawVersion + "\n")
    for _, r := range CanonicalLaw {
        sb.WriteString(fmt.Sprintf("%s|%s|%s|%s\n", r.Role, r.Port, r.Equation, r.Level))
    }
    h := sha256.Sum256([]byte(sb.String()))
    return hex.EncodeToString(h[:])
}

func EnforceLaw(role, port string) {
    found := false
    expectedPort := ""
    for _, r := range CanonicalLaw {
        if r.Role == role {
            found = true
            expectedPort = r.Port
            break
        }
    }
    if !found {
        log.Fatalf("LAW VIOLATION: role %s not in canonical law", role)
    }
    if port != expectedPort {
        log.Fatalf("LAW VIOLATION: role %s must use port %s, got %s", role, expectedPort, port)
    }
    log.Printf("LAW OK: role=%s port=%s hash=%s", role, port, LawHash()[:16])
}
