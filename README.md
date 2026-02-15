<p align="center">
  <img src="logos/crud_jt_logo_black.png#gh-light-mode-only" alt="Logo Light" />
  <img src="logos/crud_jt_logo.png#gh-dark-mode-only" alt="Logo Dark" />
</p>

<p align="center">
  Fast, file-backed JSON token for REST APIs with multi-process support
</p>

<p align="center">
  <a href="https://www.patreon.com/crudjt">
    <img src="logos/buy_me_a_coffee_orange.svg" alt="Buy Me a Coffee"/>
  </a>
</p>

## Why?  
[Escape the JWT trap: predictable login, safe logout](https://medium.com/@CoffeeMainer/jwt-trap-login-logout-under-control-7f4495d6024d)

CRUDJT runs a small local coordinator inside your app.
One process acts as a leader, all others talk to it

## In short

CRUDJT gives you stateful sessions without JWT pain and without distributed complexity

# Installation

```sh
go get github.com/crudjt/crudjt-go/v1
```

## How to use

- One process starts the master
- All other processes connect to it

## Start CRUDJT master (once)

Start the CRUDJT master when your application boots  

Only **one process** should do this  
The master is responsible for session state and coordination  

### Generate an encrypted key

```sh
export CRUDJT_ENCRYPTED_KEY=$(openssl rand -base64 48)
```

```go
import (
  "github.com/crudjt/crudjt-go/v1"
  "os"
)

crudjt.StartMaster(crudjt.ServerConfig	{
  EncryptedKey: os.Getenv("CRUDJT_ENCRYPTED_KEY"),
  StoreJtPath: "path/to/local/storage", // optional
  Host: "127.0.0.1", // default
  Port: 50051, // default
})
```
The encrypted key must be the same for all processes

## Connect to an existing CRUDJT master

Use this in all other processes  

Typical examples:
- multiple local processes
- background jobs
- forked processes

```go
import "github.com/crudjt/crudjt-go/v1"

crudjt.ConnectToMaster(crudjt.ClientConfig	{
  Host: "127.0.0.1", // default
  Port: 50051, // default
})
```

### Process layout

App boot  
 ├─ Process A → start_master  
 ├─ Process B → connect_to_master  
 └─ Process C → connect_to_master  

# C

```go
data := map[string]interface{}{"user_id": 42, "role": 11} // required
ttl := 3600 * 24 * 30 // optional: token lifetime (seconds)

// Optional: read limit
// Each read decrements the counter
// When it reaches zero — the token is deleted
silence_read := 10

token, error := crudjt.Create(&data, &ttl, &silence_read)
// token, error == HBmKFXoXgJ46mCqer1WXyQ <nil>
```

```go
data := map[string]interface{}{"user_id": 42, "role": 11}

// To disable token expiration or read limits, pass `nil`
crudjt.Create(
  &data,
  nil, // disable TTL
  nil // disable read limit
)
```

# R

```go
result, error := crudjt.Read("HBmKFXoXgJ46mCqer1WXyQ")
// result, error == map[metadata:map[ttl:101001 silence_read:9] data:map[user_id:42 role:11]] <nil>
```

```go
// When expired or not found token
result, error := crudjt.Read("HBmKFXoXgJ46mCqer1WXyQ")
// result, error ==  map[] <nil>
```

# U

```go
data := map[string]interface{}{"user_id": 42, "role": 11}
// `nil` disables limits
ttl := 600
silence_read := 100

result, error := crudjt.Update("HBmKFXoXgJ46mCqer1WXyQ", data, ttl, silence_read)
// result, error == true <nil>
```

```go
// When expired or not found token
result, error := crudjt.Update("HBmKFXoXgJ46mCqer1WXyQ", nil, nil)
// result, error == false <nil>
```

# D
```go
result, error := crudjt.Delete("HBmKFXoXgJ46mCqer1WXyQ")
// result, error == true <nil>
```

```go
// when expired/not found token
result, error := crudjt.Delete("HBmKFXoXgJ46mCqer1WXyQ")
// result, error == false <nil>
```

# Performance
**40k** requests of **256 bytes** — median over 10 runs  
ARM64 (Apple M1+), macOS 15.6.1  
Go 1.23.0/1.24.1

| Function | CRUDJT (Go) | JWT (Go) | redis-session-store (Ruby, Rails 8.0.4) |
|----------|-------|------|------|
| C        | 0.387 second | 0.196 second ⭐ | 4.057 seconds |
| R        | `0.022 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | 0.235 second | 7.011 seconds |
| U        | `0.479 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | X | 3.49 seconds |
| D        | `0.247 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | X | 6.589 seconds |

[Full benchmark results](https://github.com/exwarvlad/benchmarks)

# Storage (File-backed)  
Backed by a disk-based B-tree for predictable reads, writes, and deletes

## Disk footprint  
**40k** tokens of **256 bytes** each — median over 10 creates  
darwin23, APFS  

`48 MB`  

[Full disk footprint results](https://github.com/Cm7B68NWsMNNYjzMDREacmpe5sI1o0g40ZC9w1y/disk_footprint)

## Path Lookup Order
Stored tokens are placed in the **file system** according to the following order

1. Explicitly set via `crudjt.StartMaster(crudjt.ServerConfig { StoreJtPath: "/custom/path/to/file_system_db",})`
2. Default system location
   - **Linux**: `/var/lib/store_jt`
   - **macOS**: `/usr/local/var/store_jt`
3. Project root directory (fallback)

## Storage Characteristics
* CRUDJT **automatically removing expired tokens** after start and every 24 hours without blocking the main thread   
* **Storage automatically fsyncs every 500ms**, meanwhile tokens ​​are available from cache

# Multi-process Coordination
For multi-process scenarios, CRUDJT uses gRPC over an insecure local port for same-host communication only. It is not intended for inter-machine or internet-facing usage

# Limits
The library has the following limits and requirements

- **Go version:** tested with 1.24.1
- **Supported platforms:** Linux, macOS (x86_64 / arm64)
- **Maximum json size per token:** 256 bytes
- **`encrypted_key` format:** must be Base64
- **`encrypted_key` size:** must be 32, 48, or 64 bytes

# Contact & Support
<p align="center">
  <img src="logos/crud_jt_logo_favicon_black_160.png#gh-light-mode-only" alt="Visit Light" />
  <img src="logos/crud_jt_logo_favicon_white_160.png#gh-dark-mode-only" alt="Visit Dark" />
</p>

- **Custom integrations / new features / collaboration**: support@crudjt.com  
- **Library support & bug reports:** [open an issue](https://github.com/crudjt/crudjt-go/issues)


# Lincense
CRUDJT is released under the [MIT License](LICENSE.txt)

<p align="center">
  💘 Shoot your g . ? Love me out via <a href="https://www.patreon.com/crudjt">Patreon Sponsors</a>!
</p>
