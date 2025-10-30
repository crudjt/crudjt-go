<p align="center">
  <img src="logos/crud_jt_logo_black.png#gh-light-mode-only" alt="Logo Light" />
  <img src="logos/crud_jt_logo.png#gh-dark-mode-only" alt="Logo Dark" />
</p>

<p align="center">
  Simplifies user session. Login/Logout/Authorization
</p>

<p align="center">
  <a href="https://www.patreon.com/exwarvlad">
    <img src="logos/buy_me_a_coffee_orange.svg" alt="Buy Me a Coffee"/>
  </a>
</p>

# Installation

```sh
go get github.com/crudjt/crudjt-go/v1
```

Import and configure CRUD JT

```go
import "github.com/crudjt/crudjt-go/v1"

// # openssl rand -base64 48 # In your terminal
// => your_encrypted_base64/48
crudjt.Start(crudjt.Config{
  EncryptedKey: "your_encrypted_base64/32/48/64",
  StoreJtPath: "your_path_to_file_storage"
})
```

# C

```go
crudjt.Create(&map[string]interface{}{"user_id": 42, "role": 11}, nil, nil)
=> HBmKFXoXgJ46mCqer1WXyQ <nil>
```

```go
// with ttl — token time-to-live in seconds
ttl := 3600 * 24 * 30

crudjt.Create(&map[string]interface{}{"user_id": 42, "role": 11}, ttl, nil)
=> HBmKFXoXgJ46mCqer1WXyQ <nil>
```

```go
☕ = 🐰🥚
```

# R

```go
// ...
crudjt.Read("HBmKFXoXgJ46mCqer1WXyQ")
=> map[data:map[user_id:42 role:11]] <nil>
```

```go
// with ttl
crudjt.Read("HBmKFXoXgJ46mCqer1WXyQ")
=> map[metadata:map[ttl:3] data:map[user_id:42 role:11]] <nil>

// after 1 second
crudjt.read("HBmKFXoXgJ46mCqer1WXyQ")
=> map[metadata:map[ttl:2] data:map[user_id:42 role:11]] <nil>

// still second
crudjt.read("HBmKFXoXgJ46mCqer1WXyQ")
=> map[metadata:map[ttl:1] data:map[user_id:42 role:11]] <nil>

// ups
crudjt.read("HBmKFXoXgJ46mCqer1WXyQ")
=> map[] <nil>
```

```go
// with 🐰🥚
```

# U

```go
crudjt.Update("HBmKFXoXgJ46mCqer1WXyQ", &map[string]interface{}{"user_id": 42, "role": 8})
=> true <nil> // map[data:map[user_id:42 role:8]]
```

```go
// supported ttl update
ttl := 41

CRUD_JT.update("HBmKFXoXgJ46mCqer1WXyQ", &map[string]interface{}{"user_id": 42, "role": 8}, ttl, nil)
=> true <nil> // map[metadata:map[ttl:41] data:map[user_id:42 role:8]] <nil>
```

```go
// supported 🐰🥚 update
```

```go
// when expired/not found token
crudjt.update("HBmKFXoXgJ46mCqer1WXyQ", &map[string]interface{}{"user_id": 42, "role": 8})
=> false <nil>
```

# D
```go
// when token exist
crudjt.Delete("HBmKFXoXgJ46mCqer1WXyQ")
=> true <nil>
```

```go
// when expired/not found token
crudjt.Delete("HBmKFXoXgJ46mCqer1WXyQ")
=> false <nil>
```

# Performance
**40k** requests of **256 bytes** — median over 10 runs  
ARM64 (Apple M1+), macOS 15.6.1  
Go 1.23.0/1.24.1

| Function | CRUD JT (Go) | JWT (Go) | redis-session-store (Ruby, Rails 8.0.4) |
|----------|-------|------|------|
| C        | 0.387 second | 0.196 second ⭐ | 4.057 seconds |
| R        | `0.022 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | 0.235 second | 7.011 seconds |
| U        | `0.479 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | X | 3.49 seconds |
| D        | `0.247 second` ![Logo Favicon Light](logos/crud_jt_logo_favicon_white.png#gh-light-mode-only) ![Logo Favicon Dark](logos/crud_jt_logo_favicon_black.png#gh-dark-mode-only) | X | 6.589 seconds |

[Full results](https://github.com/exwarvlad/benchmarks)

# Storage (Store JT)

## Path Lookup Order
Stored tokens are placed in the **file system** according to the following order

1. Explicitly set via `crudjt.Start(StoreJtPath: "/custom/path/to/file_system_db", ...)`
2. Default system location
   - **Linux**: `/var/lib/store_jt`
   - **macOS**: `/usr/local/var/store_jt`
3. Project root directory (fallback)

## Storage Characteristics
* Store JT **automatically removing expired tokens** every 24 hours without blocking the main thread   
* **Store JT automatically fsyncs every 500ms**, meanwhile tokens ​​are available from cache
* Store JT is available for one process to open per instance for the time being

## Configuration

You can configure the library before starting it

```go
import "github.com/crudjt/crudjt-go/v1"

// Required configuration
crudjt.Start(crudjt.Config{
  EncryptedKey: "your_encrypted_base64/32/48/64"
})

// Optional configuration
crudjt.Start(crudjt.Config{
  EncryptedKey: "your_encrypted_base64/32/48/64",
  StoreJtPath: "/custom/path/to/file_storage_db"
})
```

`crudjt.Start(config)`  
Initializes the CRUD JT process and opens the Store JT using the given configuration  
Must be called before performing any operations  

###### Configuration options (`map[string]any`)  

`EncryptedKey: string`  
Specifies the encrypted key (in Base64 format)  
**Required**

`StoreJtPath: string`  
Overrides the default File DB storage path  
**Optional**

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
CRUD JT is released under the [MIT License](LICENSE.txt)

<p align="center">
  💘 Shoot your g . ? Love me out via <a href="https://www.patreon.com/exwarvlad">Github Sponsors</a>!
</p>
