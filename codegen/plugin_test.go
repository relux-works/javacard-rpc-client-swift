package codegen_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/javacard-rpc-client-swift/codegen"
	"github.com/relux-works/javacard-rpc/pluginapi"
)

type parityFixture struct {
	Schema  *pluginapi.Schema
	Options pluginapi.Options
	Files   []pluginapi.File
	Error   string
}

// The public API reproduces independently checked v0.4.5 bytes, filenames,
// adapter file order and refusal text for the four pinned fixtures.
func TestPluginReleasedBaselineParity(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.json")
	if err != nil || len(paths) != 4 {
		t.Fatalf("fixtures: %v %v", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var f parityFixture
			if err := json.Unmarshal(data, &f); err != nil {
				t.Fatal(err)
			}
			files, err := codegen.Plugin{}.Generate(f.Schema, f.Options)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != f.Error || !reflect.DeepEqual(files, f.Files) {
				t.Fatalf("baseline drift: error %q want %q; ordered file bytes equal=%v", message, f.Error, reflect.DeepEqual(files, f.Files))
			}
			// Options belonging to other backends must not alter Swift output.
			for _, memory := range []string{"", "clear_on_deselect", "clear_on_reset"} {
				for _, sim := range []string{"", "com.klinec:jcardsim:3.0.5.9", "works.relux:jcardsim:3.0.5.9-relux.1"} {
					o := f.Options
					o.StreamMemory = memory
					o.SimulatorDependency = sim
					got, e := codegen.Plugin{}.Generate(f.Schema, o)
					if fmt.Sprint(e) != fmt.Sprint(err) || !reflect.DeepEqual(got, files) {
						t.Fatal("irrelevant options changed Swift output")
					}
				}
			}
		})
	}
}

func schema() *pluginapi.Schema {
	return &pluginapi.Schema{Applet: pluginapi.Applet{Name: "Demo", AID: "A000000001", CLA: 0x80}, Methods: map[string]*pluginapi.Method{}}
}

// Stream rejection covers request/response, single/mixed payloads, and all
// supported workspace policies through Plugin.Generate, with ordinary bytes controls.
func TestPluginRejectsStreamsWithoutPartialFiles(t *testing.T) {
	for _, side := range []string{"request", "response"} {
		for _, mixed := range []bool{false, true} {
			for _, workspace := range []string{"", "transient", "persistent"} {
				t.Run(fmt.Sprintf("%s/mixed=%v/workspace=%s", side, mixed, workspace), func(t *testing.T) {
					s := schema()
					s.Applet.StreamWorkspace = workspace
					fields := []pluginapi.Field{{Name: "payload", Type: pluginapi.FieldTypeStream, MaxLength: 1024, ChunkSize: 64, Location: pluginapi.ParameterLocationData}}
					if mixed {
						fields = append([]pluginapi.Field{{Name: "prefix", Type: pluginapi.FieldTypeU8}}, fields...)
					}
					m := &pluginapi.Method{Name: "transfer", INS: 0x20}
					s.Methods[m.Name] = m
					msg := &pluginapi.Message{Fields: fields}
					if side == "request" {
						m.Request = msg
					} else {
						m.Response = msg
					}
					files, err := codegen.Plugin{}.Generate(s, pluginapi.Options{})
					want := `method "transfer" ` + side + ` field "payload": unsupported field type "stream"`
					if err == nil || err.Error() != want || files != nil {
						t.Fatalf("stream admitted or partial files: %v %v", files, err)
					}
					msg.Fields[len(msg.Fields)-1].Type = pluginapi.FieldTypeBytes
					if files, err := (codegen.Plugin{}).Generate(s, pluginapi.Options{}); err != nil || len(files) != 2 {
						t.Fatalf("ordinary bytes control: %v", err)
					}
				})
			}
		}
	}
}

// Errors are returned before any package files; this samples renderer invariants,
// rather than claiming that the backend validates the facade's complete schema contract.
func TestPluginRejectsInvalidRendererInputs(t *testing.T) {
	cases := []struct {
		name   string
		change func(*pluginapi.Schema)
		want   string
	}{
		{"aid", func(s *pluginapi.Schema) { s.Applet.AID = "ZZ" }, "decode applet AID"},
		{"nil-method", func(s *pluginapi.Schema) { s.Methods["broken"] = nil }, `method "broken" is nil`},
		{"fixed-request", func(s *pluginapi.Schema) {
			s.Methods["broken"] = &pluginapi.Method{Name: "broken", Request: &pluginapi.Message{Fields: []pluginapi.Field{{Name: "data", Type: pluginapi.FieldTypeBytesFixed}}}}
		}, "fixed-length bytes field must have length > 0"},
	}
	if files, err := (codegen.Plugin{}).Generate(nil, pluginapi.Options{}); err == nil || err.Error() != "schema is nil" || files != nil {
		t.Fatalf("nil: %v %v", files, err)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := schema()
			c.change(s)
			files, err := codegen.Plugin{}.Generate(s, pluginapi.Options{})
			if err == nil || !strings.Contains(err.Error(), c.want) || files != nil {
				t.Fatalf("%v %v", files, err)
			}
		})
	}
	if files, err := (codegen.Plugin{}).Generate(schema(), pluginapi.Options{}); err != nil || len(files) != 2 {
		t.Fatalf("valid control: %v", err)
	}
}

// In-memory generation returns independently owned outputs, keeps the input
// schema unchanged, and creates no files on either success or refusal.
func TestPluginInMemoryOwnership(t *testing.T) {
	dir := t.TempDir()
	prior, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(prior)
	s := schema()
	before, _ := json.Marshal(s)
	first, err := codegen.Plugin{}.Generate(s, pluginapi.Options{Namespace: "Custom.swift"})
	if err != nil {
		t.Fatal(err)
	}
	first[0].Data[0] ^= 1
	second, err := codegen.Plugin{}.Generate(s, pluginapi.Options{Namespace: "Custom.swift"})
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(first, second) {
		t.Fatal("output buffers alias")
	}
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatal("input mutated")
	}
	codegen.Plugin{}.Generate(nil, pluginapi.Options{})
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("filesystem effects: %v %v", entries, err)
	}
	if second[0].Name != "Package.swift" || second[1].Name != "Sources/DemoClient/DemoClient.swift" {
		t.Fatal("file order or products changed")
	}
}
