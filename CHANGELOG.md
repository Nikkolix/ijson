# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [1.0.0] - 2025-09-27

Initial stable release of a small Go library for generic, discriminator-based polymorphic unmarshaling for both JSON and MessagePack.

### Added
- Polymorphic decoding into interface fields using discriminators.
- Works with `encoding/json` and `vmihailenco/msgpack`.
- Three strategies:
  - Registry-based via `ijson.RDecodable[I, D]` and `ijson.RegisterT[T, I, D](disc)` / `ijson.Register[I, D](disc, factory)`.
  - External decider via `ijson.Decodable[I, D, Decider]` where `Decider.Decide(D) (I, error)`.
  - Self-deciding via `ijson.XDecodable[I, X]` where `X.Decide() (I, error)`.
- Utilities: `ijson.ResetRegistries()` for clean test setup.
- Robust error handling for invalid payloads, missing registry, unknown types, and decider errors.

### Installation
- `go get github.com/Nikkolix/ijson@v1.0.0`

### Compatibility
- Go 1.23+ recommended.

### Notes
- First stable API. No breaking changes from prior tags.

