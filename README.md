<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="logos/crudjt_logo_white_on_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="logos/crudjt_logo_dark_on_white.svg">
    <img alt="Shows a dark logo" src="logos/crudjt_logo_dark.png">
  </picture>
    </br>
    Go SDK for the fast, file-backed, scalable JSON token engine
</p>

<p align="center">
  <a href="https://www.patreon.com/crudjt">
    <img src="logos/buy_me_a_coffee_orange.svg" alt="Buy Me a Coffee"/>
  </a>
</p>

> ⚠️ Version 1.0.0-beta — production testing phase   
> API is stable. Feedback is welcome before the final 1.0.0 release

Fast B-tree–backed token store for stateful user sessions  
Provides authentication and authorization across multiple processes  
Optimized for vertical scaling on a single server  

# Installation

```sh
go get github.com/crudjt/crudjt-go/v1
```

## How to use

- One process starts the master
- All other processes connect to it

## Start CRUDJT master (once)

Start the CRUDJT master when your application boots

Only **one process** can do this for a **single token storage**  

The master process manages sessions and coordination    
All functions can also be used directly from it

### Generate a new secret key (terminal)

```sh
export CRUDJT_SECRET_KEY=$(openssl rand -base64 48)
```

### Start master (go)

```go
import (
  "github.com/crudjt/crudjt-go/v1"
  "os"
)

crudjt.StartMaster(crudjt.ServerConfig	{
  SecretKey: os.Getenv("CRUDJT_SECRET_KEY"),
  StoreJtPath: "path/to/local/storage", // optional
  Host: "127.0.0.1", // default
  Port: 50051, // default
})
```
*Important: Use the same `secret_key` across all sessions. If the key changes, previously stored tokens cannot be decrypted and will return `nil` or `false`*  

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
// To disable token expiration or read limits, pass `nil`
crudjt.Create(
  &map[string]interface{}{"user_id": 42, "role": 11},
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
> Metrics will be published after 1.0.0-beta GitHub Actions builds

# Storage (File-backed)  

## Disk footprint  
> Metrics will be published after 1.0.0-beta GitHub Actions builds

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
- **`secret_key` format:** must be Base64
- **`secret_key` size:** must be 32, 48, or 64 bytes

# Contact & Support
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="logos/crudjt_favicon_160x160_white_on_dark.svg" width=160 height=160>
    <source media="(prefers-color-scheme: light)" srcset="logos/crudjt_favicon_160x160_dark_on_white.svg" width=160 height=160>
    <img alt="Shows a dark favicon in light color mode and a white one in dark color mode" src="logos/crudjt_favicon_160x160_white.png" width=160 height=160>
  </picture>
</p>

- **Custom integrations / new features / collaboration**: support@crudjt.com  
- **Library support & bug reports:** [open an issue](https://github.com/crudjt/crudjt-go/issues)


# Lincense
CRUDJT is released under the [MIT License](LICENSE.txt)

<p align="center">
  💘 Shoot your g . ? Love me out via <a href="https://www.patreon.com/crudjt">Patreon Sponsors</a>!
</p>
