# Contributing to **puremail**

First of all, **thank you for taking the time to contribute!** Your help, ideas and bug reports make this project better for everyone.

> **Any kind of improvement is welcome –** new features, refactorings, docs, tests – as long as your merge request is neat, clearly explains **what changed** and **what advantages** it brings.

---

## 1. Getting Started

1. **Fork** the repository and clone your fork locally.
2. Ensure you have **Go ≥ 1.22** installed.
3. Install development tools:

   ```bash
   go install golang.org/x/tools/cmd/goimports@latest  # formatting
   go install honnef.co/go/tools/cmd/staticcheck@latest # extra lint
   ```
4. Run the full test suite to confirm a clean slate:

   ```bash
   go test ./...
   ```

## 2. Branching & Workflow

| Step                 | What to do                                                                   |
| -------------------- | ---------------------------------------------------------------------------- |
| **feature / bugfix** | Create a branch named `feat/<slug>` or `fix/<slug>` off `main`.              |
| **commit**           | Commit early and often; keep changes focused.                                |
| **PR**               | Open a pull request against `main` once the feature is ready and tests pass. |

We follow the “fork & pull request” model—direct pushes to `main` are disabled.

## 3. Code Style

* Run `goimports -w .` **before each commit**.
* `go vet ./...` and `staticcheck ./...` should return **zero issues**.
* Keep public APIs documented.

## 4. Commit Messages

Use the [Conventional Commits](https://www.conventionalcommits.org/) spec:

```
<type>(<scope>): <subject>

<body>  # optional, wrap at 72 chars

<footer> # optional
```

Examples:

* `feat(parser): add punycode auto‑decode`
* `fix(hash): correct output length under 20 bytes`

## 5. Pull‑Request Checklist

* [ ] **Clear title**  — imperative mood, no trailing period.
* [ ] Description covers **what changed**, **why it matters**, and **benefits**.
* [ ] Linked issue(s) or reference to discussion.
* [ ] `go test ./...` passes locally.
* [ ] Added/updated unit‑tests and benchmarks where relevant.
* [ ] Updated **README.md** / **CHANGELOG.md** if behaviour changed.
* [ ] CI checks are green.

> *Tip: keep PRs under **400 LOC** – large PRs take longer to review.*

## 6. Testing & Benchmarks

* Add **table‑driven tests** for new edge‑cases.
* Maintain or improve overall code coverage.
* For performance‑critical paths (`Parse()`, `Bytes()`), include `go test -bench` results in the PR description.

## 7. Documentation

If you add a new public function or change existing behaviour, update the relevant docs:

* **README.md** usage examples
* Go‑doc comments (`godoc` output)
* Any diagrams or markdown docs in `/docs`

## 8. Code of Conduct

This project adopts the [Contributor Covenant v2.0](CODE_OF_CONDUCT.md). Be respectful and inclusive in all interactions.

## 9. Security Policy

Please **do not** open public issues for security vulnerabilities. Instead, mail `git@sunsung.fun`.

## 10. License

By submitting a contribution you agree that your code will be released under the project’s existing ** Apache License 2.0**.
