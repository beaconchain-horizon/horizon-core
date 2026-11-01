# 📜 Horizon Core — Project Constitution (قانون اساسی)

**WARNING:** Any AI Agent or Developer modifying this codebase MUST read and strictly adhere to the following laws. Violating these laws is strictly forbidden and will result in immediate rejection of code.

## 🏛️ The Three Absolute Laws (قوانین سه‌گانه)

1.  **Law #1: License Enforced Everywhere (قانون لایسنس اجباری)**
    *   **Rule:** ALL internal and external communications (Nodes, Switches, APIs, Agents) MUST be authenticated using a valid License.
    *   **Forbidden:** No public endpoint (except `/health`) is allowed without `X-License-ID` validation. Bypassing the `LicenseAuthMiddleware` for any reason (even for benchmarks) is a direct violation.
2.  **Law #2: Billion-Scale Readiness (آمادگی برای یک میلیارد لایسنس)**
    *   **Rule:** The database and infrastructure must be optimized to handle up to 1,000,000,000 (1 Billion) unique agent licenses. Bulk inserts must use batching (`CreateInBatches`).
3.  **Law #3: Zero-Storage of Private Keys (ممنوعیت مطلق کلید خصوصی)**
    *   **Rule:** Private keys, Seeds, or Mnemonics MUST NEVER be hardcoded, stored in databases, or written to disk.
    *   **Allowed:** Only Public Keys and Signatures may be stored. Private keys MUST be generated in volatile RAM (in-memory) during runtime.

## 🏗️ Architectural Rules (قوانین معماری)

4.  **Law #4: Crypto Sovereignty (حاکمیت رمزنگاری)**
    *   **Rule:** The native cryptographic engine is `secp256k1`. It is independent of US/NIST mandates.
    *   **Rule:** PQC (Post-Quantum) is a plug-in option, never a mandatory dependency. The `Crypto-Agility` layer must remain intact.
5.  **Law #5: Branch Isolation (جدایی شاخه‌ها)**
    *   **Rule:** NEVER merge branches blindly (e.g., `git merge mine`). Code must be cherry-picked or explicitly checked out (`git checkout mine -- <file>`). 
6.  **Law #6: Liara Purge & Docker Safety (پاکسازی لیارا و ایمنی داکر)**
    *   **Rule:** All configurations, workflows, and references to "Liara" MUST be deleted.
    *   **Rule:** Do NOT delete `Dockerfile` or `docker-compose.yml`. Simply replace the word "Liara" with "Horizon" inside them.
7.  **Law #7: Anonymity in Legal Docs (گمنامی در اسناد حقوقی)**
    *   **Rule:** The name "Mahdi Amolimoghadam" or any personal name MUST NOT appear in public legal documents (GDPR, SCC). Use "Horizon Core Architecture Team" instead.

---
*By executing code in this repository, you automatically agree to be bound by these Constitutional Laws.*
