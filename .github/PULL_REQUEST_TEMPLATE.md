<!--
13 Layers to Production — merge gate.
Score every APPLICABLE layer. Mark N/A with a reason (a local backend has no CDN, etc.).
The Opus merge reviewer must not approve while an applicable, critical layer is ❌ missing.
Security (8) and Auth (4) are near-always critical.
-->

## What & why
<!-- One or two sentences: what this change does and the problem it solves. -->

## Verified locally
<!-- Commands run + result — evidence, not assertions (tests, build, manual exercise). -->

## 13-Layer Production Readiness
✅ addressed · 🟡 partial · ❌ missing · ⬜ n/a (give a reason). Only applicable layers gate.

| # | Layer | Status | Note |
|---|-------|:---:|------|
| 1 | Frontend foundations | | |
| 2 | API & backend logic | | |
| 3 | Database & storage | | |
| 4 | Auth & permissions | | |
| 5 | Hosting & deployment | | |
| 6 | Cloud & compute sizing | | |
| 7 | CI/CD & version control | | |
| 8 | Security & RLS | | |
| 9 | Rate limiting | | |
| 10 | Caching & CDN | | |
| 11 | Load balancing & scaling | | |
| 12 | Error tracking & logs | | |
| 13 | Availability & recovery | | |

## Rollback / blast radius
<!-- How this reverts if it goes wrong; what breaks if it fails; is it behind the auto-rollback rails? -->

## Secrets / config
<!-- New env/secrets by NAME only (STOP 21 — never values). Where set (Fly / CF / GH Actions). -->
