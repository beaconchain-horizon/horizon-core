# Horizon Core Protocol — License & Consensus Rules

## 1. Graceful Null Handling (No Crashes)
- Any field related to Customer Identity (`HardwareID`, `UserID`, `LicenseID`) may be empty during initial testing or block generation.
- The Switch and 8 Angel Nodes MUST NOT panic or crash if these fields are empty.
- They must proceed to calculate the Merkle Root and sign the block.

## 2. Merkle Root is the Sole Source of Truth
- All signatures MUST be derived strictly from the Binary/Ternary Merkle Roots.
- Ad-hoc JSON signing or custom node signatures are strictly prohibited.
- Every node signs the Merkle Root provided by the consensus payload.

## 3. License vs. Block Generation
- A Block REQUIRES a License to be generated (to maintain the chain).
- Transactions within a Block consume the License volume.
- When volume is exhausted, the block is sealed, and a new Block + License must be generated.
