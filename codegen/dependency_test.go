package codegen_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// The actual compiled backend graph contains only this target, the standard
// library and released pluginapi; no facade/parser/template package is reachable.
func TestBackendReleasedDependencyBoundary(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "-f", `{{if not .Standard}}{{.ImportPath}}{{end}}`, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dependency graph: %v %s", err, output)
	}
	for _, path := range strings.Fields(string(output)) {
		if path != "github.com/relux-works/javacard-rpc/pluginapi" && !strings.HasPrefix(path, "github.com/relux-works/javacard-rpc-client-swift/codegen") {
			t.Fatalf("forbidden backend dependency: %s", path)
		}
	}
	cmd = exec.Command("go", "list", "-m", "-json", "github.com/relux-works/javacard-rpc/pluginapi")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("API module: %v %s", err, output)
	}
	var module struct {
		Version string
		Replace json.RawMessage
	}
	if err := json.Unmarshal(output, &module); err != nil {
		t.Fatal(err)
	}
	if module.Version != "v0.1.0" || len(module.Replace) > 0 {
		t.Fatalf("API must be public v0.1.0 without replacement: %s", output)
	}
}
