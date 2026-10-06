package codegen_test

import (
	"encoding/json"
	"os/exec"
	"testing"
)

// SwiftPM's production manifest loader still exposes the original root runtime
// library product and target; generated package products are separately pinned.
func TestRuntimeProductManifest(t *testing.T) {
	cmd := exec.Command("swift", "package", "--package-path", "..", "dump-package")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("runtime manifest: %v %s", err, output)
	}
	var manifest struct {
		Name     string
		Products []struct {
			Name    string
			Targets []string
		}
		Targets []struct{ Name, Path string }
	}
	if err := json.Unmarshal(output, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "javacard-rpc-client-swift" || len(manifest.Products) != 1 {
		t.Fatalf("runtime package/products drift: %s", output)
	}
	product := manifest.Products[0]
	if product.Name != "JavaCardRPCClient" || len(product.Targets) != 1 || product.Targets[0] != "JavaCardRPCClient" {
		t.Fatalf("runtime public product drift: %s", output)
	}
	for _, target := range manifest.Targets {
		if target.Name == "JavaCardRPCClient" && target.Path == "Sources/JavaCardRPCClient" {
			return
		}
	}
	t.Fatal("runtime moved from its normal ecosystem root")
}
