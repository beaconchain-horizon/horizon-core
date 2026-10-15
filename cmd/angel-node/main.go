package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"horizon-core/internal/crypto"
)

// ─── NODE LAW v1.0 ───
// هر نود فقط و فقط یک معادله دارد. خارج از آن → crash.

var (
	role   string
	port   string
	pubHex string
)

type Req struct {
	Left            string `json:"left"`
	Right           string `json:"right"`
	Header          string `json:"header"`
	Message         string `json:"message"`
	Signature       string `json:"signature"`
	MerkleType      string `json:"merkle_type"`
	Used            int64  `json:"used"`
	Volume          int64  `json:"volume"`
	IssuedAt        int64  `json:"issued_at"`
	ExpiresAt       int64  `json:"expires_at"`
	UserID          string `json:"user_id"`
	ProductID       string `json:"product_id"`
	ExpectedUser    string `json:"expected_user"`
	ExpectedProduct string `json:"expected_product"`
}

func main() {
	role = os.Getenv("NODE_ROLE")
	port = os.Getenv("NODE_PORT")
	pubHex = os.Getenv("NODE_PUBKEY")
	if role == "" || port == "" {
		log.Fatal("FAIL-CLOSED: NODE_ROLE and NODE_PORT required")
	}

	// قانون: نقش باید در لیست مجاز باشد، وگرنه crash
	allowed := map[string]bool{
		"merkle": true, "hash": true, "ecdsa": true,
		"type": true, "volume": true, "time": true, "user": true,
	}
	if !allowed[role] {
		log.Fatalf("CRASH: role not allowed: %s", role)
	}
	// قانون: نقش ecdsa به pubkey نیاز دارد
	if role == "ecdsa" && pubHex == "" {
		log.Fatal("CRASH: ecdsa role requires NODE_PUBKEY")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"role": role, "port": port, "status": "online"})
	})
	r.POST("/verify", verifyHandler)

	log.Printf("🔹 Angel Node [%s] listening :%s", role, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("CRASH: bind failed on %s: %v", port, err)
	}
}

func verifyHandler(c *gin.Context) {
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"vote": false, "error": "bad request"})
		return
	}
	start := time.Now()

	var vote bool
	var eq string

	// ─── فقط یک معادله بر اساس نقش ───
	switch role {
	case "merkle":
		vote, eq = verifyMerkle(req)
	case "hash":
		vote, eq = verifyHash(req)
	case "ecdsa":
		vote, eq = verifyECDSA(req)
	case "type":
		vote, eq = verifyType(req)
	case "volume":
		vote, eq = verifyVolume(req)
	case "time":
		vote, eq = verifyTime(req)
	case "user":
		vote, eq = verifyUser(req)
	default:
		// هرگز نباید اینجا برسیم — قانون crash
		log.Fatalf("CRASH: role drift detected: %s", role)
	}

	c.JSON(200, gin.H{
		"role":       role,
		"vote":       vote,
		"equation":   eq,
		"latency_ms": time.Since(start).Milliseconds(),
	})
}

// ─── معادله‌ها: هر کدام فقط یک کار ───

func verifyMerkle(r Req) (bool, string) {
	if r.Left == "" && r.Right == "" {
		return false, "SHA256(left||right): empty"
	}
	h := sha256.New()
	h.Write([]byte(r.Left))
	h.Write([]byte(r.Right))
	return true, "SHA256(left||right)=" + hex.EncodeToString(h.Sum(nil))[:16]
}

func verifyHash(r Req) (bool, string) {
	if r.Header == "" {
		return false, "SHA256(header): empty"
	}
	h := sha256.Sum256([]byte(r.Header))
	return true, "SHA256(header)=" + hex.EncodeToString(h[:])[:16]
}

func verifyECDSA(r Req) (bool, string) {
	if r.Signature == "" || r.Message == "" {
		return false, "secp256k1.Verify: empty"
	}
	pub, err := crypto.Secp256k1PublicKeyFromHex(pubHex)
	if err != nil {
		log.Fatalf("CRASH: pubkey parse failed: %v", err)
	}
	ok, err := crypto.VerifyData(pub, []byte(r.Message), r.Signature)
	if err != nil {
		return false, "secp256k1.Verify err: " + err.Error()
	}
	return ok, "secp256k1.Verify=" + boolStr(ok)
}

func verifyType(r Req) (bool, string) {
	if r.MerkleType != "binary" && r.MerkleType != "ternary" {
		return false, "type ∉ {binary,ternary}"
	}
	return true, "type=" + r.MerkleType
}

func verifyVolume(r Req) (bool, string) {
	if r.Volume <= 0 {
		return false, "volume<=0"
	}
	if r.Used >= r.Volume {
		return false, "used>=volume (exhausted)"
	}
	return true, "used<volume"
}

func verifyTime(r Req) (bool, string) {
	now := time.Now().Unix()
	if now < r.IssuedAt {
		return false, "now<issued_at"
	}
	if now >= r.ExpiresAt {
		return false, "now>=expires_at"
	}
	return true, "issued≤now<expires"
}

func verifyUser(r Req) (bool, string) {
	if r.ExpectedUser == "" || r.ExpectedProduct == "" {
		return false, "expected_* empty"
	}
	if r.UserID != r.ExpectedUser {
		return false, "user_id mismatch"
	}
	if r.ProductID != r.ExpectedProduct {
		return false, "product_id mismatch"
	}
	return true, "user/product match"
}

func boolStr(b bool) string {
	if b { return "OK" }
	return "FAIL"
}
