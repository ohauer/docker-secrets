package config

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzConfigLoad(f *testing.F) {
	// Seed with valid YAML
	f.Add([]byte(`secretStore:
  address: "http://localhost:8200"
  authMethod: "token"
  token: "test"
secrets:
  - name: "test"
    key: "secret/data/test"
    mountPath: "secret"
    kvVersion: "v2"
    template:
      data:
        test: '{{ .value }}'
    files:
      - path: "/tmp/test"
        mode: "0600"
`))

	// Seed with namespace
	f.Add([]byte(`secretStore:
  address: "http://localhost:8200"
  namespace: "team-a"
  authMethod: "token"
  token: "test"
secrets:
  - name: "test"
    key: "secret/data/test"
    mountPath: "secret"
    namespace: "team-b"
    kvVersion: "v2"
    refreshInterval: "30m"
    template:
      data:
        test: '{{ .value }}'
    files:
      - path: "/tmp/test"
        mode: "0600"
`))

	// Seed with edge cases
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`---`))

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Skip()
		}

		// Should not panic
		_, _ = Load(configPath)
	})
}

func FuzzValidateFilePath(f *testing.F) {
	// Seed with valid paths
	f.Add("/tmp/test")
	f.Add("/var/lib/secrets/test")

	// Seed with attack vectors
	f.Add("../../../etc/passwd")
	f.Add("/tmp/../../../etc/passwd")
	f.Add("/tmp/./test")
	f.Add("relative/path")
	f.Add("/tmp/test\x00hidden")
	f.Add("/tmp/test\n/etc/passwd")

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic
		_ = validateFilePath(path)
	})
}
