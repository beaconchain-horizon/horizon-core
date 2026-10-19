#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

echo "=============================================="
echo " Horizon Core - Merkle v2 RFC 6962 Upgrade"
echo "=============================================="

ALLOWED=(
  "cmd/switch/switch.go"
  "cmd/switch/admin.go"
  "cmd/switch/license_sign.go"
  "cmd/switch/db.go"
  "internal/merkle/merkle.go"
)

echo "[1/8] Checking allowed files exist..."
for f in "${ALLOWED[@]}"; do
  test -f "$f" || { echo "ERROR: missing $f"; exit 1; }
done

echo "[2/8] Checking go.mod (must stay at go 1.24.0)..."
go version
if ! grep -q '^go 1\.24\.0$' go.mod; then
  echo "ERROR: go.mod is not exactly 'go 1.24.0'"
  head -5 go.mod
  exit 1
fi
echo "OK: go.mod stays at 1.24.0"

echo "[3/8] Creating backup..."
BACKUP=".merkle-v2-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP"
for f in "${ALLOWED[@]}"; do
  mkdir -p "$BACKUP/$(dirname "$f")"
  cp "$f" "$BACKUP/$f"
done
echo "Backup: $BACKUP"

echo "[4/8] Writing RFC 6962 Merkle implementation..."
cat > internal/merkle/merkle.go <<'GOEOF'
package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// RFC 6962 Merkle Tree
// leaf   = SHA256(0x00 || data)
// parent = SHA256(0x01 || left || right)

const (
	leafPrefix = byte(0x00)
	nodePrefix = byte(0x01)
)

type Node struct {
	Hash  string
	Left  *Node
	Right *Node
}

type Tree struct {
	Root   *Node
	Leaves []*Node
}

type ProofStep struct {
	Hash   string
	IsLeft bool
}

func LeafHash(data []byte) string {
	h := sha256.New()
	h.Write([]byte{leafPrefix})
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func NodeHash(leftHex, rightHex string) (string, error) {
	left, err := hex.DecodeString(leftHex)
	if err != nil {
		return "", fmt.Errorf("invalid left hash: %w", err)
	}
	right, err := hex.DecodeString(rightHex)
	if err != nil {
		return "", fmt.Errorf("invalid right hash: %w", err)
	}
	if len(left) != sha256.Size || len(right) != sha256.Size {
		return "", fmt.Errorf("hashes must be 32 bytes")
	}
	h := sha256.New()
	h.Write([]byte{nodePrefix})
	h.Write(left)
	h.Write(right)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func NewTree(data [][]byte) *Tree {
	tree := &Tree{}
	if len(data) == 0 {
		return tree
	}
	tree.Leaves = make([]*Node, len(data))
	for i, d := range data {
		tree.Leaves[i] = &Node{Hash: LeafHash(d)}
	}
	tree.Root = buildRFC6962(tree.Leaves)
	return tree
}

func buildRFC6962(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	k := largestPow2LessThan(len(nodes))
	left := buildRFC6962(nodes[:k])
	right := buildRFC6962(nodes[k:])
	hash, err := NodeHash(left.Hash, right.Hash)
	if err != nil {
		panic(err)
	}
	return &Node{Hash: hash, Left: left, Right: right}
}

func largestPow2LessThan(n int) int {
	if n <= 1 {
		return 0
	}
	k := 1
	for (k << 1) < n {
		k <<= 1
	}
	return k
}

func (t *Tree) GetRootHash() string {
	if t == nil || t.Root == nil {
		sum := sha256.Sum256(nil)
		return hex.EncodeToString(sum[:])
	}
	return t.Root.Hash
}

func (t *Tree) GenerateProof(index int) []ProofStep {
	if t == nil || index < 0 || index >= len(t.Leaves) {
		return nil
	}
	var proof []ProofStep
	var walk func(nodes []*Node, idx int)
	walk = func(nodes []*Node, idx int) {
		if len(nodes) <= 1 {
			return
		}
		k := largestPow2LessThan(len(nodes))
		if idx < k {
			right := buildRFC6962(nodes[k:])
			proof = append(proof, ProofStep{Hash: right.Hash, IsLeft: false})
			walk(nodes[:k], idx)
			return
		}
		left := buildRFC6962(nodes[:k])
		proof = append(proof, ProofStep{Hash: left.Hash, IsLeft: true})
		walk(nodes[k:], idx-k)
	}
	walk(t.Leaves, index)
	return proof
}

func VerifyProof(leafHash string, proof []ProofStep, rootHash string) bool {
	current := leafHash
	for _, step := range proof {
		var next string
		var err error
		if step.IsLeft {
			next, err = NodeHash(step.Hash, current)
		} else {
			next, err = NodeHash(current, step.Hash)
		}
		if err != nil {
			return false
		}
		current = next
	}
	return current == rootHash
}

func VerifyProofData(data []byte, proof []ProofStep, rootHash string) bool {
	return VerifyProof(LeafHash(data), proof, rootHash)
}

func RootFromData(data [][]byte) string {
	return NewTree(data).GetRootHash()
}

func RootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		return NewTree(nil).GetRootHash()
	}
	data := make([][]byte, len(hashes))
	for i, v := range hashes {
		decoded, err := hex.DecodeString(v)
		if err != nil {
			return ""
		}
		data[i] = decoded
	}
	return RootFromData(data)
}

// LegacyRootFromHashes — Horizon v1 Merkle (blocks #0-#3)
func LegacyRootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		h := sha256.Sum256([]byte("empty"))
		return hex.EncodeToString(h[:])
	}
	nodes := make([]string, len(hashes))
	copy(nodes, hashes)
	for len(nodes) > 1 {
		next := make([]string, 0, (len(nodes)+1)/2)
		for i := 0; i < len(nodes); i += 2 {
			if i+1 >= len(nodes) {
				next = append(next, nodes[i])
				continue
			}
			combined := append([]byte(nodes[i]), []byte(nodes[i+1])...)
			hash := sha256.Sum256(combined)
			next = append(next, hex.EncodeToString(hash[:]))
		}
		nodes = next
	}
	return nodes[0]
}
GOEOF
echo "OK: merkle.go written"

echo "[5/8] Patching switch.go (Node.js)..."

node - <<'JSEOF'
const fs = require('fs');
const p = 'cmd/switch/switch.go';
let s = fs.readFileSync(p, 'utf8');
let n = 0;

// Block struct — add MerkleAlgo
if (!s.includes('MerkleAlgo string')) {
    s = s.replace(
        /(MerkleRoot string\s+`json:"merkle_root"`)/,
        '$1\n\tMerkleAlgo string    `json:"merkle_algo"`'
    );
    n++;
    console.log('  + Block.MerkleAlgo');
}

// simpleMerkleRoot — dual algorithm
const oldRoot = /func simpleMerkleRoot\(txIDs \[\]string\) string \{[\s\S]*?\n\}/;
const newRoot = `func simpleMerkleRoot(txIDs []string, algo string) string {
	if len(txIDs) == 0 {
		return hashData([]byte("empty"))
	}
	if algo == "v2" {
		data := make([][]byte, len(txIDs))
		for i, id := range txIDs {
			data[i] = []byte(id)
		}
		return merkle.RootFromData(data)
	}
	hashes := make([]string, len(txIDs))
	for i, id := range txIDs {
		hashes[i] = hashData([]byte(id))
	}
	return merkle.LegacyRootFromHashes(hashes)
}`;
if (oldRoot.test(s)) {
    s = s.replace(oldRoot, newRoot);
    n++;
    console.log('  + simpleMerkleRoot (dual)');
}

// Genesis — add MerkleAlgo: "v1"
if (!s.includes('MerkleAlgo: "v1"')) {
    s = s.replace(
        /(MerkleRoot:\s*hashData\(\[\]byte\("genesis"\)\),)/,
        '$1\n\t\t\t\tMerkleAlgo: "v1",'
    );
    n++;
    console.log('  + Genesis.MerkleAlgo');
}

// mineBlock — use algo based on height
const oldMine = /merkleRoot := simpleMerkleRoot\(txIDs\)/;
const newMine = `merkleAlgo := "v1"
	if last.BlockNum+1 >= 4 {
		merkleAlgo = "v2"
	}
	merkleRoot := simpleMerkleRoot(txIDs, merkleAlgo)`;
if (oldMine.test(s)) {
    s = s.replace(oldMine, newMine);
    n++;
    console.log('  + mineBlock algo selection');
}

// Block creation — add MerkleAlgo field
s = s.replace(
    /(MerkleRoot:\s*merkleRoot,)(\s*\n\s*TxCount:)/,
    '$1\n\t\tMerkleAlgo: merkleAlgo,$2'
);
n++;

// CORS — add localhost:8000
if (!s.includes('localhost:8000')) {
    s = s.replace(
        /"http:\/\/localhost:8080"/,
        '"http://localhost:8080", "http://localhost:8000"'
    );
    n++;
    console.log('  + CORS localhost:8000');
}

// saveLicenseHandler — populate merkle_root
if (!s.includes('licenseMerkleRoot(lic)')) {
    const oldSave = /(lic := &License\{[\s\S]*?Status:\s*"active",\s*\n\t\tHardwareID:\s*getHardwareIDOrEmpty\(\),\s*\n\t\})/;
    if (oldSave.test(s)) {
        s = s.replace(oldSave, '$1\n\n\tif lic.MerkleRoot == "" {\n\t\tlic.MerkleRoot = licenseMerkleRoot(lic)\n\t}');
        n++;
        console.log('  + saveLicenseHandler merkle');
    }
}

fs.writeFileSync(p, s);
console.log('  Total patches in switch.go:', n);
JSEOF

echo "[6/8] Patching admin.go..."

node - <<'JSEOF'
const fs = require('fs');
const p = 'cmd/switch/admin.go';
let s = fs.readFileSync(p, 'utf8');
let n = 0;

if (!s.includes('licenseMerkleRoot(lic)')) {
    const old = /(lic := &License\{[\s\S]*?Status:\s*"active",\s*\n\t\})/;
    if (old.test(s)) {
        s = s.replace(old, '$1\n\n\tif lic.MerkleRoot == "" {\n\t\tlic.MerkleRoot = licenseMerkleRoot(lic)\n\t}');
        n++;
        console.log('  + admin license merkle');
    }
}

fs.writeFileSync(p, s);
console.log('  Total patches in admin.go:', n);
JSEOF

echo "[7/8] Patching license_sign.go..."

node - <<'JSEOF'
const fs = require('fs');
const p = 'cmd/switch/license_sign.go';
let s = fs.readFileSync(p, 'utf8');
let n = 0;

// Add import
if (!s.includes('horizon-core/internal/merkle')) {
    s = s.replace(
        /"net\/http"\n\t"time"/,
        '"net/http"\n\t"strconv"\n\t"time"\n\n\t"horizon-core/internal/merkle"'
    );
    n++;
    console.log('  + import merkle');
}

// Add licenseMerkleRoot helper before licenseCanonicalMessage
if (!s.includes('func licenseMerkleRoot')) {
    const helper = `// licenseMerkleRoot builds the RFC 6962 root for license fields.
func licenseMerkleRoot(lic *License) string {
	leaves := [][]byte{
		[]byte("license_id=" + lic.LicenseID),
		[]byte("user_id=" + lic.UserID),
		[]byte("product_id=" + lic.ProductID),
		[]byte("volume=" + strconv.Itoa(lic.Volume)),
		[]byte("duration=" + strconv.Itoa(lic.Duration)),
		[]byte("issued_at=" + strconv.FormatInt(lic.IssuedAt, 10)),
		[]byte("expires_at=" + strconv.FormatInt(lic.ExpiresAt, 10)),
		[]byte("hardware_id=" + lic.HardwareID),
	}
	return merkle.RootFromData(leaves)
}

// licenseCanonicalMessage returns the canonical JSON payload`;

    s = s.replace('// licenseCanonicalMessage returns the canonical JSON payload', helper);
    n++;
    console.log('  + licenseMerkleRoot helper');
}

fs.writeFileSync(p, s);
console.log('  Total patches in license_sign.go:', n);
JSEOF

echo "[8/8] Formatting, building, pushing..."

gofmt -w \
  cmd/switch/switch.go \
  cmd/switch/admin.go \
  cmd/switch/license_sign.go \
  internal/merkle/merkle.go

# Scope check
CHANGED="$(git diff --name-only)"
echo "Changed files:"
echo "$CHANGED"

while IFS= read -r f; do
  [ -z "$f" ] && continue
  case "$f" in
    cmd/switch/switch.go|cmd/switch/admin.go|cmd/switch/license_sign.go|cmd/switch/db.go|internal/merkle/merkle.go) ;;
    *) echo "ERROR: unauthorized change: $f"; exit 1 ;;
  esac
done <<< "$CHANGED"

if git diff --name-only | grep -qE '^(go\.mod|go\.sum)$'; then
  echo "ERROR: go.mod/go.sum modified"
  exit 1
fi

# Build
go build -o switch.exe ./cmd/switch
echo "OK: build passed"
ls -lh switch.exe

# Commit
git add cmd/switch/switch.go cmd/switch/admin.go cmd/switch/license_sign.go internal/merkle/merkle.go
git commit -m "feat(merkle): RFC 6962 v2 with v1 archive for blocks 0-3"
git push origin main

echo ""
echo "=============================================="
echo " DONE"
echo "=============================================="
echo " Blocks #0-#3 : v1 (archive)"
echo " Blocks #4+   : v2 (RFC 6962)"
echo " Licenses     : auto merkle_root"
echo " CORS         : localhost:8000 allowed"
echo " Go           : go.mod = 1.24.0"
