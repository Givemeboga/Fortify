# Security Policy

Fortify is a web-application **security scanner**. This policy covers vulnerabilities
**in Fortify itself** — not the findings Fortify reports *about* your targets.

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

### What to expect

Fortify is maintained by a single developer, so these are good-faith targets, not guarantees:

| Stage | Target |
|---|---|
| Acknowledgement | within 3 days |
| Initial assessment / triage | within 7 days |
| Fix or mitigation plan | by severity and complexity |

You'll be credited in the published advisory unless you prefer to remain anonymous.

## Scope

**In scope** — vulnerabilities in Fortify's own code, for example:

- SSRF / request forgery through the scan-target URL
- Injection, path traversal, or auth bypass in the API or dashboard
- Secret or API-key leakage (e.g. a provider key exposed by the backend)
- Stored or reflected XSS in the dashboard

**Out of scope**

- Findings Fortify reports about *your* targets — that's the tool working as intended
- Vulnerabilities in third-party dependencies (report upstream; do tell us if Fortify's
  particular use of one is exploitable)
- Misuse of the tool against systems you don't own (see Responsible use)
- Issues requiring physical access or an already-compromised host

## Safe harbor

We support good-faith security research. If you make a sincere effort to follow this policy,
we will not pursue or support legal action against you for your research, and we'll work with
you to understand and resolve the issue. In return, please: test only against your **own**
instance, avoid privacy violations and service disruption, and give us reasonable time to
respond before any public disclosure.

## Responsible use

Fortify's active scanner sends real injection and traversal payloads, and is intended for
**authorized testing only** — scan systems you own or have explicit written permission to
test (see [Responsible Use](README.md#responsible-use)). Misuse of the tool is out of scope
for this policy and is solely the user's responsibility.
