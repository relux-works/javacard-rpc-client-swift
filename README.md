# javacard-rpc-client-swift

Swift package backend and client runtime for [javacard-rpc](https://github.com/relux-works/javacard-rpc),
an RPC framework for Java Card smart-card applets. The backend generates typed
async clients from the javacard-rpc IDL. The Swift runtime provides APDU commands,
responses, packing helpers and a TCP transport.

## Repository layout

The repository is one Swift target plugin: root `go.mod` declares
`github.com/relux-works/javacard-rpc-client-swift`, and `codegen/` owns its Go
backend, renderer and package templates. It depends only on released
`github.com/relux-works/javacard-rpc/pluginapi v0.1.0` and the Go standard
library. Root `Package.swift` and `Sources/JavaCardRPCClient/` retain the
Swift runtime and public `JavaCardRPCClient` product. A repository tag versions
both ecosystems at the same commit.

The facade composes `codegen.Plugin{}` through `pluginapi.Plugin`. Its
`Generate` method consumes a validated schema and returns ordered in-memory
files: `Package.swift`, then `Sources/<Applet>Client/<Applet>Client.swift`.
The facade owns IDL parsing, validation and output writes. Ordinary generation
preserves the signed core v0.4.5 compatibility baseline; Swift stream fields
continue to return an error without partial package files. Facade composition
and release delivery are handled in the facade repository.

## Installation (Swift Package Manager)

```swift
.package(url: "https://github.com/relux-works/javacard-rpc-client-swift.git", exact: "0.2.2")
```

```swift
.product(name: "JavaCardRPCClient", package: "javacard-rpc-client-swift")
```

Version `0.2.2` uses the shared root tag `v0.2.2`. The matching Go backend pin is
`github.com/relux-works/javacard-rpc-client-swift v0.2.2` (import
`github.com/relux-works/javacard-rpc-client-swift/codegen`). Both resolve to
the same repository commit. See [release notes](RELEASE_NOTES.md) for the
changes and compatibility limits.

## Usage

Define your interface in the javacard-rpc TOML IDL and run its codegen. The generated
Swift package exposes typed async methods and its own applet transport protocol.
The preserved v0.4.5 output imports Foundation and is standalone; using this
runtime's `APDUTransport` requires an adapter to that generated protocol.
See the [javacard-rpc](https://github.com/relux-works/javacard-rpc) repository
for the IDL reference, code generation, and an end-to-end example.

## Tools and validation

- Go 1.24+: `go build ./...`, `go vet ./...`, and `go test -count=1 -v ./...`
  validate the backend, public dependency boundary, pinned output fixtures and
  generated Swift client behavior. Generated-client checks require Swift 6.2+.
- Swift/SPM: `swift test` runs Swift Testing tests in
  `Tests/JavaCardRPCClientTests/`; normal SPM artifacts go to `.build/`.
- Xcode: `xcodebuild -scheme javacard-rpc-client-swift -destination
  'generic/platform=iOS Simulator' -derivedDataPath .temp/DerivedData build
  CODE_SIGNING_ALLOWED=NO` builds the iOS runtime. Use an available concrete
  iOS Simulator destination and `test` to execute the same tests on iOS.
  Some SPM build products may also appear in `DerivedData/`; move these
  task-local outputs into `.temp/` after the command exits.
- Behavioral controls: `go run .scripts/verify-mutants.go --out
  .temp/codegen-mutants` runs narrowing mutants in isolated copies and records
  full logs there. `--only fixed-31-token-preserved` selects the Swift execution
  attack that preserves the original guard text while admitting length 31.

See `codegen/testdata/README.md` for fixture provenance and evidence limits.
Task-specific command logs and parity reports live under `.temp/`.

<!-- relux-ecosystem:start -->

## About Relux Works

This project is part of the open-source ecosystem of
[Relux Works](https://relux.works), an AI-native software development studio.
We build fixed-price MVPs, rescue vibe-coded apps, run local AI inference, and
train teams to work with coding agents. Much of the infrastructure behind that
work is open source.

- Full catalog: [relux.works/en/open-source](https://relux.works/en/open-source/)
- Agentic enablement: [agent harnesses & team training](https://relux.works/en/agentic-enablement/)
- Hire us the agent-native way: point your assistant at `https://api.relux.works/mcp`
- Contact: ivan@relux.works

<!-- relux-ecosystem:end -->

## License

See [LICENSE](LICENSE).
