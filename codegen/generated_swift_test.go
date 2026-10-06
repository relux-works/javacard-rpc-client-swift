package codegen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/relux-works/javacard-rpc-client-swift/codegen"
	"github.com/relux-works/javacard-rpc/pluginapi"
)

// Execute the generated public client with a recording transport: short/long
// fixed-byte inputs must throw before transmission, while exact bytes transmit.
// This behavioral witness supplements the donor's source-snippet assertions.
func TestGeneratedSwiftFixedBytesRejectBeforeTransmit(t *testing.T) {
	if _, err := exec.LookPath("swift"); err != nil {
		t.Fatal("Swift toolchain required for generated client validation")
	}
	s := schema()
	for i, name := range []string{"setHash", "pack"} {
		fields := []pluginapi.Field{{Name: "hash", Type: pluginapi.FieldTypeBytesFixed, FixedLength: 32, Location: pluginapi.ParameterLocationData}}
		if name == "pack" {
			fields = append([]pluginapi.Field{{Name: "prefix", Type: pluginapi.FieldTypeU8}}, fields...)
		}
		s.Methods[name] = &pluginapi.Method{Name: name, INS: byte(i + 1), Request: &pluginapi.Message{Fields: fields}}
	}
	files, err := codegen.Plugin{}.Generate(s, pluginapi.Options{})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, f := range files {
		path := filepath.Join(root, filepath.FromSlash(f.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, f.Data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	// The generated manifest's product stays unchanged; this temporary consumer
	// adds its own Swift Testing target rather than changing production templates.
	manifest := `// swift-tools-version: 6.2
import PackageDescription
let package = Package(name: "GeneratedConsumer", products: [.library(name: "DemoClient", targets: ["DemoClient"])], targets: [.target(name: "DemoClient"), .testTarget(name: "ConsumerTests", dependencies: ["DemoClient"])])
`
	if err := os.WriteFile(filepath.Join(root, "Package.swift"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	tests := `import Foundation
import Testing
@testable import DemoClient
actor Recorder: DemoTransport {
 var calls = 0
 func transmit(cla: UInt8, ins: UInt8, p1: UInt8, p2: UInt8, data: Data?) async throws -> (sw: UInt16, data: Data) {
  calls += 1
  return (0x9000, Data())
 }
}
// Wrong sizes reject through generated client methods without calling transport.
@Test(arguments: [31, 33]) func generatedFixedLengthRejection(count: Int) async throws {
 let recorder = Recorder()
 let client = DemoClient(transport: recorder)
 do {
  try await client.setHash(hash: Data(repeating: 1, count: count))
  Issue.record("Single wrong size admitted")
 } catch { #expect(String(describing: error) == "invalidResponse") }
 #expect(await recorder.calls == 0)
 do {
  try await client.pack(prefix: 7, hash: Data(repeating: 1, count: count))
  Issue.record("Mixed wrong size admitted")
 } catch { #expect(String(describing: error) == "invalidResponse") }
 #expect(await recorder.calls == 0)
 try await client.setHash(hash: Data(repeating: 1, count: 32))
 try await client.pack(prefix: 7, hash: Data(repeating: 1, count: 32))
 #expect(await recorder.calls == 2)
}
`
	path := filepath.Join(root, "Tests/ConsumerTests")
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "GeneratedTests.swift"), []byte(tests), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("swift", "test", "--package-path", root)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated Swift consumer: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}
