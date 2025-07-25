[![Go Report Card](https://goreportcard.com/badge/github.com/Bookshelf-Writer/puremail)](https://goreportcard.com/report/github.com/Bookshelf-Writer/puremail)

![GitHub repo file or directory count](https://img.shields.io/github/directory-file-count/Bookshelf-Writer/puremail?color=orange)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/Bookshelf-Writer/puremail?color=green)
![GitHub repo size](https://img.shields.io/github/repo-size/Bookshelf-Writer/puremail)

# puremail

A **zero‑allocation**, high‑throughput Go library for *strict* e‑mail parsing, tag trimming, binary
serialisation and DNS‑MX probing.  
The parser normalises case, removes disposable `+` / `=` tags **before** validation, caches its own
results and lets you hash or encode an address in a single line.

> **Focus:** production back‑ends that need predictable latency and memory
> footprint. Exhaustive RFC‑5322 edge‑cases are intentionally ignored.

---

## Features

| ✔                                   | Description                                                                 |
|-------------------------------------|-----------------------------------------------------------------------------|
| **Prefix trimming**                 | `bob+promo=gophers@gmail.com` → `bob@gmail.com` (prefixes kept internally). |
| **RFC‑ish validation**              | Login & domain checked against a pragmatic subset of the RFC.               |
| **Parser cache**                    | Same address parsed only once thanks to `singleflight`; toggle via config.  |
| **MX probing with smart cache**     | `HasMX()` uses a sharded, TTL‑aware cache with concurrency limits.          |
| **CRC‑protected bytes**             | `Bytes()` / `Decode()` round‑trip with CRC‑32 guard.                        |
| **BLAKE2b‑160 hashes**              | `Hash()` (login+domain) & `HashFull()` (including prefixes).                |
| **100 % allocation‑free fast path** | All hot methods avoid heap use.                                             |
| **Fuzz‑tested & benchmarked**       | >500 k/s parse on a single core (see `go test -bench .`).                   |

---

## Installation

```bash
go get github.com/Bookshelf-Writer/puremail
```

---

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/Bookshelf-Writer/puremail"
)

func main() {
	// Initialize with default configuration
	puremail.InitDefault()

	addr, err := puremail.New("Alice+dev=go@example.io")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(addr.Mail())        // alice@example.io
	fmt.Println(addr.MailFull())    // alice+dev=go@example.io
	fmt.Printf("%x\n", addr.Hash()) // 20‑byte BLAKE2b‑160
}
```

The package can be configured with a custom ConfigObj:

```go
package main

import (
	"context"
	"time"

	"github.com/Bookshelf-Writer/puremail"
)

func main() {
	config := puremail.ConfigObj{
		NoCache: false,
		MX: puremail.ConfigMxObj{
			TllPos:       12 * time.Hour,
			TllNeg:       30 * time.Minute,
			RefreshAhead: 20 * time.Minute,

			TimeoutDns:      500 * time.Millisecond,
			TimeoutDnsBurst: 3 * time.Second,
			TimeoutRefresh:  60 * time.Second,

			ShardAbs:     8,
			ShardMaxSize: 20_000,

			ConcurrencyLimitLookupMX: 500,
		},
		Ctx: context.Background(),
	}

	puremail.Init(config)

	// Use the package functions...
}
```

---

## Configuration (`ConfigObj`)

| Field       | Type              | Purpose / default                                                            |
|-------------|-------------------|------------------------------------------------------------------------------|
| **NoCache** | `bool`            | `true` disables the internal *singleflight* cache used by `New` / `NewFast`. |
| **MX**      | `ConfigMxObj`     | Nested object that tunes the MX resolver cache (see below).                  |
| **Ctx**     | `context.Context` | Root context for background goroutines. Defaults to `context.Background()`.  |

### `ConfigMxObj`

| Field                      | Default  | What it does                                                       |
|----------------------------|----------|--------------------------------------------------------------------|
| `TllPos`                   | `6h`     | TTL for *positive* MX answers.                                     |
| `TllNeg`                   | `15m`    | TTL for *negative* answers (NXDOMAIN / no records).                |
| `RefreshAhead`             | `10m`    | Time **before** TTL when an entry may be refreshed asynchronously. |
| `TimeoutDns`               | `400ms`  | Hard limit for a single DNS lookup.                                |
| `TimeoutDnsBurst`          | `2s`     | Upper bound when many lookups queue at once.                       |
| `TimeoutRefresh`           | `90s`    | How often the cleaner scans & evicts expired items.                |
| `ShardAbs`                 | `4`      | log₂ of cache shards ⇒ `2⁴ = 16` shards (1 .. 31).                 |
| `ShardMaxSize`             | `10 000` | Max entries per shard (oldest drop first).                         |
| `ConcurrencyLimitLookupMX` | `250`    | Global semaphore guarding parallel DNS queries.                    |

> Call `puremail.Init(&cfg)` once at program start.
> Calling nothing is identical to `puremail.InitDefault()`.

---

## Constructors

| Function            | Behaviour                                                           |
|---------------------|---------------------------------------------------------------------|
| `New(s string)`     | Validates and **trims prefixes** (`+`, `=`).                        |
| `NewFast(s string)` | Same validation, but prefixes are treated as normal chars (faster). |

---

## API reference

### `EmailObj` methods

| Method       | Returns            | Comment                                                    |
|--------------|--------------------|------------------------------------------------------------|
| `Login()`    | `string`           | Local part without prefixes.                               |
| `Domain()`   | `string`           | Domain in lower‑case.                                      |
| `Prefixes()` | `[]EmailPrefixObj` | Slice of preserved prefixes.                               |
| `Mail()`     | `string`           | Canonical `<login>@<domain>`.                              |
| `MailFull()` | `string`           | Original address with prefixes.                            |
| `String()`   | `string`           | Debug representation.                                      |
| `Bytes()`    | `[]byte`           | Binary payload + CRC‑32.                                   |
| `Hash()`     | `[20]byte`         | BLAKE2b‑160 of login+domain.                               |
| `HashFull()` | `[20]byte`         | Same, but includes prefixes.                               |
| `HasMX()`    | `error`            | `nil` if at least one MX exists. Cached, concurrency‑safe. |

### `EmailPrefixObj`

| Method     | Purpose                         |
|------------|---------------------------------|
| `String()` | Original text (`"dev"`).        |
| `Prefix()` | Delimiter char (`'+'` / `'='`). |

### Stand‑alone helpers

| Function         | Use case                                 |
|------------------|------------------------------------------|
| `Decode([]byte)` | Recreate `EmailObj` from `Bytes()` blob. |

---

## Usage examples

```go
addr, _ := puremail.New("bob+promo=gophers@gmail.com")

// 1. Basic fields
fmt.Println(addr.Login())        // bob
fmt.Println(addr.Domain()) // gmail.com
fmt.Println(addr.Mail()) // bob@gmail.com
fmt.Println(addr.MailFull()) // bob+promo=gophers@gmail.com
fmt.Println(addr.String()) // [ 'bob@gmail.com', ['+promo', '=gophers'] ]

// 2. Prefix enumeration
for _, p := range addr.Prefixes() {
fmt.Printf("tag %c = %s\n", p.Prefix(), p.String())
}

// 3. Hashes
fmt.Printf("stable hash  : %x\n", addr.Hash())
fmt.Printf("hash w/tags  : %x\n", addr.HashFull())

// 4. Binary round‑trip
blob := addr.Bytes()
back, _ := puremail.Decode(blob)
fmt.Println(back.Mail()) // bob@gmail.com

// 5. MX check (cached)
if err := addr.HasMX(); err != nil {
log.Printf("domain has no MX: %v", err)
}

// 6. NewFast: keep prefixes
fast, _ := puremail.NewFast("bob+promo=gophers@gmail.com")
fmt.Println(fast.MailFull()) // unchanged
```

---

## Encoding / decoding in detail

```go
e, _ := puremail.New("alice+dev=go@example.io")
payload := e.Bytes() // safe to store in Redis or pass over the wire
again, err := puremail.Decode(payload)
if err != nil { panic(err) }
```

The format is:

```
<len(login)><login><len(domain)><domain>[ <tag><len(txt)><txt> ... ]<crc‑32LE>
```

Any corruption (or truncated payload) is caught by the CRC check.

---

## MX cache life‑cycle

```
┌─parse.HasMX()──────────────────┐
│ shard lookup (CRC‑32 hash)     │  ← constant‑time
│ ├─ fresh? → return             │
│ └─ group.Do(domain, dnsQuery)  │  ← singleflight + semaphore
└────────────────────────────────┘
```

* Positive TTL (`TllPos`) and negative TTL (`TllNeg`) are fully configurable.
* A background goroutine prunes expired entries every `TimeoutRefresh`.
* Cache size is bounded per shard; oldest keys are dropped.

---

## Limitations

* ASCII input only; supply punycode yourself (`пример.укр` → `xn--e1afmkfd.xn--j1amh`).
* No quoted‑local‑part, comments or IP‑literals.
* Max total length **254 bytes**.
* `HasMX()` issues network DNS lookups (honours context cancellation).

---

---

### Mirrors

- https://git.bookshelf-writer.fun/Bookshelf-Writer/puremail