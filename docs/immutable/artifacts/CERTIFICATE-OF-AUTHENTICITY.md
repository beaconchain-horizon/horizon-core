# HORIZON CORE — Certificate of Authenticity

> **Document type:** Public attestation
> **Issuer:** Horizon Core protocol authority
> **Version:** v1.4
> **Genesis locked:** 2026-11-03
> **Chain ID:** `horizon-core-v1.4`

---

## 1. Declaration

The undersigned authority hereby certifies that the cryptographic artifacts
referenced in this document constitute the **immutable genesis state** of the
Horizon Core v1.4 protocol.

These artifacts — chain identity, genesis block hash, block #1 hash, and the
authority public key — are the **single source of truth** for any party
wishing to verify the authenticity of this protocol.

Any future mutation of these values is, by definition, a fork, and shall not
be considered a continuation of Horizon Core v1.4.

---

## 2. Canonical Values

| Field | Value | Verified in |
|:------|:------|:------------|
| Chain ID | `horizon-core-v1.4` | `GENESIS-public.json` |
| Genesis Hash | `8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603` | `GENESIS-public.json` |
| Block #1 Hash | `f981652e6c25945a2655b83f7536524441422bdeb7b9742a30a5256cd637ceba` | `GENESIS-public.json` |
| Authority Pubkey | `04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214` | `GENESIS-public.json` |
| Curve | `secp256k1` (SEC 2 v2) | `HORIZON-CORE-v1.4-STANDARDS.md §1` |
| Hash | `SHA-256` (FIPS 180-4) | same |
| Signature scheme | `ECDSA` + RFC 6979 + Low-S (BIP-62) | same |

---

## 3. Signatures

The genesis and block #1 are signed by the authority key above.

| Artifact | Signature (R‖S, 64 bytes) | Status |
|:---------|:--------------------------|:-------|
| Genesis | _partial: `afac005b…d236face`_ | ⚠ full 128-hex pending publication |
| Block #1 | _partial: `19f1b567…94bd335`_ | ⚠ full 128-hex pending publication |

> **Verification note:** the hash of block #1 is independently verifiable
> against the genesis artifact. Full ECDSA signature verification requires
> the complete 128-hex R‖S value; that value is being published separately
> and, once available, will be inserted here without altering the hashes or
> pubkey above.

---

## 4. Independent Verification Procedure

Any third party may verify these artifacts **without trusting this document
or the repository's history**:

```bash
git clone https://github.com/beaconchain-horizon/horizon-core.git
cd horizon-core/docs/immutable/artifacts
./verify-genesis.sh
[OK]   GENESIS-public.json checksum matches
[OK]   authority_pubkey matches GENESIS.json
[OK]   genesis_hash matches GENESIS.json
=== ALL CHECKS PASSED ===
این **Block #1 Hash** است — در آرتیفکت فعلی هم همین را داریم (`f981652e...6cd637ceba`). کامل ثبت شده. ✅

اما برای `ecdsa.Verify` واقعی، من به **امضای ۱۲۸ کاراکتری R‖S** نیاز دارم، نه هش. جستجو می‌کنم شاید نسخهٔ کامل در `GENESIS.json` یا `CHAIN_STATE.json` باشد — و همان‌جا `CERTIFICATE-OF-AUTHENTICITY.md` را هم می‌سازم.

```bash
cd ~/Beaconchain/horizon-core

echo "═══ ۱) جستجوی امضای کامل در فایل‌های آرتیفکت ═══"
for f in GENESIS.json CHAIN_STATE.json docs/immutable/artifacts/GENESIS-public.json; do
  [ -f "$f" ] && echo "--- $f ---" && grep -oE '[0-9a-fA-F]{128}' "$f" | head -3
done

echo ""
echo "═══ ۲) ساخت CERTIFICATE-OF-AUTHENTICITY.md ═══"
cat > docs/immutable/artifacts/CERTIFICATE-OF-AUTHENTICITY.md <<'CERT_EOF'
# HORIZON CORE — Certificate of Authenticity

> **Document type:** Public attestation
> **Issuer:** Horizon Core protocol authority
> **Version:** v1.4
> **Genesis locked:** 2026-11-03
> **Chain ID:** `horizon-core-v1.4`

---

## 1. Declaration

The undersigned authority hereby certifies that the cryptographic artifacts
referenced in this document constitute the **immutable genesis state** of the
Horizon Core v1.4 protocol.

These artifacts — chain identity, genesis block hash, block #1 hash, and the
authority public key — are the **single source of truth** for any party
wishing to verify the authenticity of this protocol.

Any future mutation of these values is, by definition, a fork, and shall not
be considered a continuation of Horizon Core v1.4.

---

## 2. Canonical Values

| Field | Value | Verified in |
|:------|:------|:------------|
| Chain ID | `horizon-core-v1.4` | `GENESIS-public.json` |
| Genesis Hash | `8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603` | `GENESIS-public.json` |
| Block #1 Hash | `f981652e6c25945a2655b83f7536524441422bdeb7b9742a30a5256cd637ceba` | `GENESIS-public.json` |
| Authority Pubkey | `04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214` | `GENESIS-public.json` |
| Curve | `secp256k1` (SEC 2 v2) | `HORIZON-CORE-v1.4-STANDARDS.md §1` |
| Hash | `SHA-256` (FIPS 180-4) | same |
| Signature scheme | `ECDSA` + RFC 6979 + Low-S (BIP-62) | same |

---

## 3. Signatures

The genesis and block #1 are signed by the authority key above.

| Artifact | Signature (R‖S, 64 bytes) | Status |
|:---------|:--------------------------|:-------|
| Genesis | _partial: `afac005b…d236face`_ | ⚠ full 128-hex pending publication |
| Block #1 | _partial: `19f1b567…94bd335`_ | ⚠ full 128-hex pending publication |

> **Verification note:** the hash of block #1 is independently verifiable
> against the genesis artifact. Full ECDSA signature verification requires
> the complete 128-hex R‖S value; that value is being published separately
> and, once available, will be inserted here without altering the hashes or
> pubkey above.

---

## 4. Independent Verification Procedure

Any third party may verify these artifacts **without trusting this document
or the repository's history**:

```bash
git clone https://github.com/beaconchain-horizon/horizon-core.git
cd horizon-core/docs/immutable/artifacts
./verify-genesis.sh
```

Expected output:
```
[OK]   GENESIS-public.json checksum matches
[OK]   authority_pubkey matches GENESIS.json
[OK]   genesis_hash matches GENESIS.json
=== ALL CHECKS PASSED ===
```

Full cryptographic verification (ECDSA verify against authority pubkey)
becomes available once the complete signatures in §3 are published.

---

## 5. Scope and Limitations

This certificate attests to the **genesis state only**.

| In scope | Out of scope |
|:---------|:-------------|
| Chain identity | Per-license issuance |
| Genesis hash | Per-block signature validity |
| Block #1 hash | Operational switch state |
| Authority pubkey | Customer deployments |

License authenticity is enforced at the switch level and verified per-request;
see `HORIZON-CORE-v1.4-SECURITY-DESIGN.md §4`.

---

## 6. Immutable Anchors

This document is anchored to the following git references:

| Reference | Value |
|:----------|:------|
| Commit | `d5fdded` |
| Tag `v1.4-artifacts` | points to `d5fdded` |
| Tag `v1.4-canonical` | earlier immutable spec commit |
| Self-fingerprint | see `CERTIFICATE-OF-AUTHENTICITY.sha256` |

---

<div align="center">
HORIZON CORE — a protocol, not a product.<br/>
Certificate of Authenticity · v1.4 · genesis locked 2026-11-03
</div>
