[![Go Report Card](https://goreportcard.com/badge/github.com/Bookshelf-Writer/puremail)](https://goreportcard.com/report/github.com/Bookshelf-Writer/puremail)

![GitHub repo file or directory count](https://img.shields.io/github/directory-file-count/Bookshelf-Writer/puremail?color=orange)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Bookshelf-Writer/puremail?color=green)
![GitHub repo size](https://img.shields.io/github/repo-size/Bookshelf-Writer/puremail)

# puremail

A blazing‑fast, zero‑allocation Go package for strict e‑mail parsing, tag trimming and
binary serialisation.  
It removes disposable `+` / `=` tags _before_ validation, normalises case, verifies
domain labels and lets you hash or encode the address in one line of code.

## Features

| ✔                       | Description                                                                      |
|-------------------------|----------------------------------------------------------------------------------|
| **Prefix trimming**     | `bob+promo=gophers@gmail.com` → `bob@gmail.com` (prefixes kept if you need them) |
| **RFC‑ish validation**  | Login & domain checked against a reduced but practical subset                    |
| **MX probing**          | Optional `HasMX()` check to see if the domain has MX records                     |
| **CRC‑protected bytes** | `Bytes()` / `Decode()` round‑trip with CRC‑32 guard                              |
| **BLAKE2b‑160 hashes**  | `Hash()` (login+domain) and `HashFull()` (including prefixes)                    |

> ⚠️ The library targets production back‑ends that need speed, not exhaustive RFC-5322 coverage.
> See *Limitations* below.

## Installation

```bash
go get github.com/Bookshelf-Writer/puremail
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/Bookshelf-Writer/puremail"
)

func main() {
	addr, err := puremail.New("Alice+dev=go@example.io")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(addr.Mail())     // alice@example.io
	fmt.Println(addr.MailFull()) // alice+dev=go@example.io
	fmt.Printf("%x\n", addr.Hash())
}
```

## API Reference

| Constructor                            | Description                                                                |
|----------------------------------------|----------------------------------------------------------------------------|
| `New(s string) (*EmailObj, error)`     | Full validation, **trims prefixes** but keeps them in the object.          |
| `NewFast(s string) (*EmailObj, error)` | Same as `New` but skips prefix parsing (`+`, `=` treated as normal chars). |

| Method                   | Returns                            | Example                                     |
|--------------------------|------------------------------------|---------------------------------------------|
| `(*EmailObj) Login()`    | `string`                           | `e.Login() // "alice"`                      |
| `(*EmailObj) Domain()`   | `string`                           | `e.Domain() // "example.io"`                |
| `(*EmailObj) Prefixes()` | `[]EmailPrefixObj`                 | `for _, p := range e.Prefixes() { ... }`    |
| `(*EmailObj) Mail()`     | login`@`domain                     | `e.Mail() // "alice@example.io"`            |
| `(*EmailObj) MailFull()` | with prefixes                      | `e.MailFull() // "alice+dev=go@example.io"` |
| `(*EmailObj) String()`   | debug string                       | `fmt.Println(e)`                            |
| `(*EmailObj) Bytes()`    | `[]byte` (CRC‑32 protected)        | `data := e.Bytes()`                         |
| `Decode(data)`           | `*EmailObj`                        | `e2, _ := puremail.Decode(data)`            |
| `(*EmailObj) Hash()`     | `[20]byte`                         | `fmt.Printf("%x", e.Hash())`                |
| `(*EmailObj) HashFull()` | `[20]byte`                         | same but with prefixes                      |
| `(*EmailObj) HasMX()`    | `error` (`nil` if at least one MX) | `if err := e.HasMX(); err != nil { ... }`   |

### `EmailPrefixObj`

| Method     | Purpose                             |
|------------|-------------------------------------|
| `String()` | original text of the prefix         |
| `Prefix()` | the delimiter char (`'+'` or `'='`) |

## Encoding/Decoding

```go
e, _ := puremail.New("bob+test=go@gmail.com")
blob := e.Bytes()

same, _ := puremail.Decode(blob)
fmt.Println(same.Mail()) // bob@gmail.com
```

## Limitations

* Only ASCII input; for non‑ASCII domains supply punycode (`пример.укр` → `xn--e1afmkfd.xn--j1amh`).
* No quoted‑local‑part, comments, IP‑literals.
* Max length **254 bytes** (same as most ESPs).
* `HasMX()` performs a network DNS lookup.

## Advantages of prefix clipping (`+`, `=`)

* **Spam control:** tags are often abused (`+noreply`, `=tracking`). Stripping before validation reduces noise.
* **Uniqueness:** your DB stores only one canonical form per user.
* **Predictable hashing:** hashes are stable even if the user modifies the tag.


---

---

### Mirrors

- https://git.bookshelf-writer.fun/Bookshelf-Writer/puremail