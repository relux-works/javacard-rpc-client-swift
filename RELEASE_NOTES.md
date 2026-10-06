# v0.2.2 — Swift backend and runtime

## Shared ecosystem pins

| Consumer | Exact pin |
| --- | --- |
| Repository root tag | `v0.2.2` |
| Go module | `github.com/relux-works/javacard-rpc-client-swift v0.2.2` |
| Go backend import | `github.com/relux-works/javacard-rpc-client-swift/codegen` |
| SwiftPM URL | `https://github.com/relux-works/javacard-rpc-client-swift.git` |
| SwiftPM version | `0.2.2` (`exact:`); root tag `v0.2.2` |
| Swift package / product | `javacard-rpc-client-swift` / `JavaCardRPCClient` |
| Backend API dependency | `github.com/relux-works/javacard-rpc/pluginapi v0.1.0` |

The released API pin originates at core commit
`ef6e04bfae8dab0f8e1ac40c9bb10713dbf09250`, tag `pluginapi/v0.1.0`;
module checksum `h1:SNT1MS/IcEzv84sJKwDGnP2j0QH8rexMpLNpOSAIlbI=`.

The root Go module and root `Package.swift` runtime share the tagged repository
commit. No nested module tag or separately versioned runtime is introduced.
Public Swift product, target, package name, source layout and supported platforms
(iOS 13+, macOS 10.15+) remain unchanged.
SwiftPM has no separate native version field to update.

## Changes and compatibility

- Adds the target-owned Go backend `codegen.Plugin{}` alongside the Swift
  runtime, depending only on the released pluginapi and the Go standard library.
- Preserves the signed core `v0.4.5` output baseline: commit
  `cfed4182356a4f4609c88f58924aac79c05ae5b6`, signed tag object
  `da77d07af5dc6866437d3db1004fdeda9738d59c`.
- Preserves unsupported Swift stream-field refusal without partial files and
  existing native runtime wire-format, status-word and fixed-byte guarantees.
- Generated clients retain their standalone Foundation-based transport
  protocol; using this runtime's `APDUTransport` still requires an adapter.

## Supported platforms and validation limits

The runtime supports iOS 13+ and macOS 10.15+. Validation covers iOS Simulator
builds and Swift Testing, macOS Swift Testing, generated Swift execution and
Go backend tests. The oldest supported OS versions, physical iOS devices and
live TCP/card connections were not exercised.
