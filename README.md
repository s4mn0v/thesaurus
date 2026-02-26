# SDK Used

[SDK Package](https://github.com/s4mn0v/bitget)

## Project Structure

```
serendipia/
├── cmd/
│   └── serendipia/
│       └── main.go             # Entry point: Initializes UI, Vault, and Trading Engine
├── internal/
│   ├── ui/
│   │   ├── layout.go           # Gocui manager and view definitions
│   │   ├── handlers.go         # UI input handling (Keybindings)
│   │   └── components/         # Reusable UI elements (sidebars, tickers)
│   ├── trading/
│   │   ├── engine.go           # High-level trading logic (Places Futures/Spot orders)
│   │   ├── futures.go          # Wrappers for SDK's MixOrderClient
│   │   └── spot.go             # Wrappers for SDK's SpotOrderClient
│   ├── security/
│   │   ├── crypto.go           # AES-GCM encryption/decryption logic
│   │   └── vault.go            # Interface to store/retrieve encrypted API keys
│   └── config/
│       └── app_config.go       # Local app settings (not the SDK config)
├── storage/
│   └── vault.db                # Encrypted local database or file
├── go.mod                      # Project dependencies
└── go.sum                      # Checksums (includes github.com/s4mn0v/bitget)
```
