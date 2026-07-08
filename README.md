# javacard-rpc-client-swift

Swift client runtime for [javacard-rpc](https://github.com/relux-works/javacard-rpc),
an RPC framework for Java Card smart-card applets. Generated clients built from the
javacard-rpc IDL use this runtime to speak the APDU protocol through ordinary typed
method calls. No manual command or response byte layout.

## Installation (Swift Package Manager)

```swift
.package(url: "https://github.com/relux-works/javacard-rpc-client-swift.git", from: "1.0.0")
```

```swift
.product(name: "JavaCardRPCClient", package: "javacard-rpc-client-swift")
```

## Usage

Define your interface in the javacard-rpc TOML IDL and run its codegen. The generated
Swift client depends on this package and exposes typed async methods for every applet
command. See the [javacard-rpc](https://github.com/relux-works/javacard-rpc) repository
for the IDL reference, code generation, and an end-to-end example.

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
