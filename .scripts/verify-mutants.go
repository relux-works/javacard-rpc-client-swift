// Run bounded behavioral controls in isolated copies, preserving the candidate.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type mutant struct{ name, file, old, new, test, tool string }

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	selected := flag.String("only", "", "one mutant, or all when empty")
	output := flag.String("out", ".temp/codegen-mutants", "task-local output directory")
	flag.Parse()
	root, err := os.Getwd()
	must(err)
	out := filepath.Join(root, *output)
	must(os.MkdirAll(out, 0755))
	apiPath, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/relux-works/javacard-rpc/pluginapi").Output()
	must(err)
	ms := []mutant{
		{"runtime-product-name", "Package.swift", ".library(name: \"JavaCardRPCClient\",", ".library(name: \"RenamedClient\",", "TestRuntimeProductManifest", "go"},
		{"pluginapi-local-copy", "go.mod", "require github.com/relux-works/javacard-rpc/pluginapi v0.1.0", "require github.com/relux-works/javacard-rpc/pluginapi v0.1.0\n\nreplace github.com/relux-works/javacard-rpc/pluginapi => " + strings.TrimSpace(string(apiPath)), "TestBackendReleasedDependencyBoundary", "go"},
		{"swift-package-byte", "codegen/package.go", "swift-tools-version: 6.2", "swift-tools-version: 6.3", "TestPluginReleasedBaselineParity", "go"},
		{"request-stream-1024", "codegen/internal/render/gen_swift.go", "for _, f := range req.Fields {", `for _, f := range req.Fields {
        if f.Type == FieldTypeStream && f.MaxLength == 1024 && f.ChunkSize == 64 { f.Type = FieldTypeBytes }`, "TestPluginRejectsStreamsWithoutPartialFiles/request", "go"},
		{"response-stream-1024", "codegen/internal/render/gen_swift.go", "field := resp.Fields[0]", `field := resp.Fields[0]
        if field.Type == FieldTypeStream && field.MaxLength == 1024 && field.ChunkSize == 64 { field.Type = FieldTypeBytes }`, "TestPluginRejectsStreamsWithoutPartialFiles/response/mixed=false", "go"},
		{"nil-schema", "codegen/plugin.go", "source, err := render.GenerateSwiftClient", `if s == nil { s = &pluginapi.Schema{Applet: pluginapi.Applet{Name: "Demo", AID: "A000000001"}} }
    source, err := render.GenerateSwiftClient`, "TestPluginRejectsInvalidRendererInputs", "go"},
		{"nil-method-broken", "codegen/internal/render/gen_swift.go", "m := entry.Method", "m := entry.Method\n        if m == nil && entry.Name == \"broken\" { m = &Method{Name: entry.Name} }", "TestPluginRejectsInvalidRendererInputs/nil-method", "go"},
		{"aid-ZZ", "codegen/internal/render/gen_swift.go", "func parseAIDBytes(aid string) ([]byte, error) {", `func parseAIDBytes(aid string) ([]byte, error) {
    if aid == "ZZ" { return []byte{0xA0,0,0,0,1}, nil }`, "TestPluginRejectsInvalidRendererInputs/aid", "go"},
		{"zero-fixed-size", "codegen/internal/render/gen_swift.go", "if f.FixedLength <= 0 {", "if f.FixedLength < 0 {", "TestPluginRejectsInvalidRendererInputs/fixed-request", "go"},
		{"fixed-31-token-preserved", "codegen/internal/render/gen_swift.go", `fmt.Sprintf("if %s.count != %d { throw TransportError.invalidResponse }", f.Name, f.FixedLength)`, `fmt.Sprintf("// if %s.count != %d { throw TransportError.invalidResponse }\n        if %s.count != %d && %s.count != 31 { throw TransportError.invalidResponse }", f.Name, f.FixedLength, f.Name, f.FixedLength, f.Name)`, "TestGeneratedSwiftFixedBytesRejectBeforeTransmit", "go"},
		{"status-9001", "Sources/JavaCardRPCClient/APDUResponse.swift", "guard sw == 0x9000 else {", "guard sw == 0x9000 || sw == 0x9001 else {", "statusWordRejection", "swift"},
		{"packer-size-1", "Sources/JavaCardRPCClient/DataPacker.swift", "guard value.count == length else {", "guard value.count == length || (length == 2 && value.count == 1) else {", "fixedBytesRejectWithoutPartialWrite", "swift"},
	}
	ran := false
	for _, m := range ms {
		if *selected != "" && *selected != m.name {
			continue
		}
		ran = true
		dir, err := os.MkdirTemp(out, m.name+"-")
		must(err)
		for _, entry := range []string{"go.mod", "go.sum", "Package.swift", "codegen", "Sources", "Tests"} {
			must(filepath.WalkDir(filepath.Join(root, entry), func(path string, d fs.DirEntry, e error) error {
				if e != nil {
					return e
				}
				rel, e := filepath.Rel(root, path)
				if e != nil {
					return e
				}
				dest := filepath.Join(dir, rel)
				if d.IsDir() {
					return os.MkdirAll(dest, 0755)
				}
				b, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				return os.WriteFile(dest, b, 0644)
			}))
		}
		path := filepath.Join(dir, m.file)
		data, err := os.ReadFile(path)
		must(err)
		if !strings.Contains(string(data), m.old) {
			panic("plant not applied: " + m.name)
		}
		must(os.WriteFile(path, []byte(strings.ReplaceAll(string(data), m.old, m.new)), 0644))
		var cmd *exec.Cmd
		if m.tool == "go" {
			cmd = exec.Command("go", "test", "-count=1", "-v", "./codegen", "-run", "^"+m.test)
		} else {
			cmd = exec.Command("swift", "test", "--filter", m.test)
		}
		cmd.Dir = dir
		data, err = cmd.CombinedOutput()
		must(os.WriteFile(filepath.Join(out, m.name+".log"), data, 0644))
		exit, ok := err.(*exec.ExitError)
		marker := "--- FAIL: " + strings.Split(m.test, "/")[0]
		if m.tool == "swift" {
			marker = "recorded an issue"
		}
		if m.name == "fixed-31-token-preserved" && !strings.Contains(string(data), "Single wrong size admitted") {
			panic("behavioral plant did not execute: " + string(data))
		}
		if !ok || exit.ExitCode() != 1 || !strings.Contains(string(data), marker) {
			panic(fmt.Sprintf("survivor/setup failure: %s err=%v\n%s", m.name, err, data))
		}
		fmt.Printf("%s: exit=%d; killed by %s; log=%s\n", m.name, exit.ExitCode(), m.test, filepath.Join(out, m.name+".log"))
	}
	if !ran {
		panic("unknown mutant: " + *selected)
	}
}
