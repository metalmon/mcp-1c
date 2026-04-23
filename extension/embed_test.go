package extension

import "testing"

func TestSourceForInstallProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		profile string
		wantErr bool
	}{
		{name: "default generic", profile: "", wantErr: false},
		{name: "generic", profile: "generic", wantErr: false},
		{name: "buh_3_0", profile: "buh_3_0", wantErr: false},
		{name: "auto rejected", profile: "auto", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, root, _, err := SourceForInstallProfile(tt.profile)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && root == "" {
				t.Fatal("expected non-empty source root")
			}
		})
	}
}
