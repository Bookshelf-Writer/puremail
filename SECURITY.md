# Security Policy

Thank you for helping us keep **puremail** and its users safe! This document describes how to report security issues and how we handle them.

## Supported Versions

The table below lists which versions of *puremail* currently receive security updates. If you are using an unsupported version, you must upgrade to a supported release before requesting a fix.

| Version                             | Status        |
| ----------------------------------- | ------------- |
| `main`                              | ✅ Supported   |
| Latest tagged release (e.g. `v1.x`) | ✅ Supported   |
| `< v1.0.0`                          | ❌ End‑of‑life |

> We generally support the two most recent minor releases. Older versions reach **end‑of‑life** 30 days after a new minor release.

## Reporting a Vulnerability

* **Private Disclosure First.** Do **not** open a public GitHub issue for security problems.
* Email **[git@sunsung.fun](mailto:git@sunsung.fun)** with the subject line **`[puremail‑SECURITY]`**.
* Please include:

  * A concise description of the vulnerability.
  * Reproduction steps or proof‑of‑concept code.
  * The version of *puremail* you are using.
  * Any relevant logs or stack traces.
* If you need to send sensitive data, request our PGP key in the initial email.

## Disclosure Policy & Timelines

| Phase                | Expected Timeframe                                                 |
| -------------------- | ------------------------------------------------------------------ |
| **Acknowledgement**  | Within **3 business days** we confirm receipt.                     |
| **Triage**           | Within **7 calendar days** we assess severity and scope.           |
| **Fix Development**  | Depending on complexity, usually **7–21 days** after triage.       |
| **Pre‑announcement** | We may coordinate an embargo with downstream integrators.          |
| **Public Release**   | A patched version, CVE (if applicable) and advisory are published. |

We strive for responsible disclosure: you are welcome to suggest an embargo period, but the final timeline is agreed mutually. Critical issues may be fast‑tracked.

## Severity Ratings

We follow the [CVSS 3.1](https://www.first.org/cvss/) standard to prioritise fixes. Low‑severity issues (CVSS < 4.0) may be deferred to the next scheduled release.

## Credit & Rewards

*We do not operate a bug‑bounty programme.* However, reporters who submit valid, non‑public vulnerabilities will be credited in the security advisory unless they request anonymity.

## Legal & License

By emailing a vulnerability report, you agree that any code or patch you submit will be licensed under the project’s existing **Apache License 2.0**.

