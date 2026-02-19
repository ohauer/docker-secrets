package config

import "testing"

func TestResolveNamespace(t *testing.T) {
	tests := []struct {
		name            string
		secretNamespace string
		globalNamespace string
		expected        string
	}{
		{
			name:            "per-secret namespace takes precedence",
			secretNamespace: "team-a",
			globalNamespace: "default",
			expected:        "team-a",
		},
		{
			name:            "use global namespace when secret has none",
			secretNamespace: "",
			globalNamespace: "default",
			expected:        "default",
		},
		{
			name:            "both empty returns empty",
			secretNamespace: "",
			globalNamespace: "",
			expected:        "",
		},
		{
			name:            "secret namespace only",
			secretNamespace: "team-b",
			globalNamespace: "",
			expected:        "team-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := &Secret{
				Namespace: tt.secretNamespace,
			}

			result := secret.ResolveNamespace(tt.globalNamespace)
			if result != tt.expected {
				t.Errorf("ResolveNamespace() = %q, want %q", result, tt.expected)
			}
		})
	}
}
