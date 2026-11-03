package main

import (
	"encoding/hex"
	"fmt"
	"os"

	k256 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	k256ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func main() {
	const pubHex = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
	const hashHex = "8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603"
	const derSigHex = "3045022100afac005b7f7565fe69ca9d58dabdf0a3914fce0854ee05f214d788e9e478b02f02207c82a05b35d1456b605a128e4074ed9d3c4e55932cc7b368516c0681d236face"

	pubB, err := hex.DecodeString(pubHex)
	if err != nil { fmt.Println("pub hex:", err); os.Exit(1) }
	pub, err := k256.ParsePubKey(pubB)
	if err != nil { fmt.Println("parse pub:", err); os.Exit(1) }

	hashB, err := hex.DecodeString(hashHex)
	if err != nil { fmt.Println("hash hex:", err); os.Exit(1) }

	sigB, err := hex.DecodeString(derSigHex)
	if err != nil { fmt.Println("sig hex:", err); os.Exit(1) }

	sig, err := k256ecdsa.ParseDERSignature(sigB)
	if err != nil { fmt.Println("parse DER sig:", err); os.Exit(2) }

	if sig.Verify(hashB, pub) {
		fmt.Println("✓ VALID — امضا با هش جنسیس و کلید authority مطابقت دارد")
		// چاپ فرمت R||S برای استفاده در دیتابیس
		fmt.Println("R||S compact hex (128 chars for authority_sig):")
		fmt.Println(hex.EncodeToString(sig.Serialize()))
	} else {
		fmt.Println("✗ INVALID — امضا با این هش/کلید سازگار نیست")
		os.Exit(3)
	}
}
