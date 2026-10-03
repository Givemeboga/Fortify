# Security Policy

Fortify is a web-application **security scanner**. This policy covers vulnerabilities
**in Fortify itself** — not findings that Fortify reports *about* your targets.

## Supported versions

Fortify is under active development; only the **latest release** receives security fixes.

| Version | Supported |
|---|---|
| latest (`main` / newest release) | ✅ |
| older releases | ❌ |

## Reporting a vulnerability

**Please do not open a public issue for security vulnerabilities.**

Report privately through GitHub's
**[private vulnerability reporting](https://github.com/Givemeboga/Fortify/security/advisories/new)**
(the repository's **Security** tab → *Report a vulnerability*). This keeps the report
confidential until a fix is available.

Please include:

- A clear description of the issue and its impact
- Steps to reproduce (a minimal proof-of-concept if possible)
- Affected version / commit
- Any suggested remediation

**What to expect.** Fortify is maintained by a single developer, so responses are
best-effort — expect an initial acknowledgement within a few days. Once confirmed, a fix
will be prioritized, and you'll be credited in the advisory unless you prefer to remain
anonymous.

## Scope & responsible use

Fortify's active scanner sends real injection and traversal payloads, and is intended for
**authorized testing only** — scan systems you own or have explicit written permission to
test (see [Responsible Use](README.md#responsible-use)). Misuse of the tool is out of scope
for this policy and is solely the user's responsibility.
