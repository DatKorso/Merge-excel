package updater

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMajor  int
		wantMinor  int
		wantPatch  int
		wantPrerel string
		wantErr    bool
	}{
		{
			name:       "Simple version with v prefix",
			input:      "v0.1.0",
			wantMajor:  0,
			wantMinor:  1,
			wantPatch:  0,
			wantPrerel: "",
			wantErr:    false,
		},
		{
			name:       "Simple version without v prefix",
			input:      "0.1.0",
			wantMajor:  0,
			wantMinor:  1,
			wantPatch:  0,
			wantPrerel: "",
			wantErr:    false,
		},
		{
			name:       "Version with alpha prerelease",
			input:      "0.1.0-alpha",
			wantMajor:  0,
			wantMinor:  1,
			wantPatch:  0,
			wantPrerel: "alpha",
			wantErr:    false,
		},
		{
			name:       "Version with v prefix and prerelease",
			input:      "v1.2.3-beta",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  3,
			wantPrerel: "beta",
			wantErr:    false,
		},
		{
			name:       "Version without patch",
			input:      "v1.2",
			wantMajor:  1,
			wantMinor:  2,
			wantPatch:  0,
			wantPrerel: "",
			wantErr:    false,
		},
		{
			name:      "Invalid version - too many parts",
			input:     "1.2.3.4",
			wantErr:   true,
		},
		{
			name:      "Invalid version - non-numeric",
			input:     "v1.x.0",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			if got.Major != tt.wantMajor {
				t.Errorf("ParseVersion() Major = %v, want %v", got.Major, tt.wantMajor)
			}
			if got.Minor != tt.wantMinor {
				t.Errorf("ParseVersion() Minor = %v, want %v", got.Minor, tt.wantMinor)
			}
			if got.Patch != tt.wantPatch {
				t.Errorf("ParseVersion() Patch = %v, want %v", got.Patch, tt.wantPatch)
			}
			if got.Prerelease != tt.wantPrerel {
				t.Errorf("ParseVersion() Prerelease = %v, want %v", got.Prerelease, tt.wantPrerel)
			}
		})
	}
}

func TestVersionIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{
			name:    "Latest is newer - patch version",
			current: "0.1.0",
			latest:  "0.1.1",
			want:    true,
		},
		{
			name:    "Latest is newer - minor version",
			current: "0.1.0",
			latest:  "0.2.0",
			want:    true,
		},
		{
			name:    "Latest is newer - major version",
			current: "0.1.0",
			latest:  "1.0.0",
			want:    true,
		},
		{
			name:    "Versions are equal",
			current: "0.1.0",
			latest:  "0.1.0",
			want:    false,
		},
		{
			name:    "Current is newer",
			current: "0.2.0",
			latest:  "0.1.0",
			want:    false,
		},
		{
			name:    "Release is newer than prerelease",
			current: "0.1.0-alpha",
			latest:  "0.1.0",
			want:    true,
		},
		{
			name:    "Prerelease is not newer than release",
			current: "0.1.0",
			latest:  "0.1.0-beta",
			want:    false,
		},
		{
			name:    "Both prereleases - same version",
			current: "0.1.0-alpha",
			latest:  "0.1.0-beta",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentVer, err := ParseVersion(tt.current)
			if err != nil {
				t.Fatalf("Failed to parse current version: %v", err)
			}

			latestVer, err := ParseVersion(tt.latest)
			if err != nil {
				t.Fatalf("Failed to parse latest version: %v", err)
			}

			got := latestVer.IsNewer(currentVer)
			if got != tt.want {
				t.Errorf("IsNewer() = %v, want %v (current: %s, latest: %s)", 
					got, tt.want, tt.current, tt.latest)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{
			name:    "v0.1.0 vs v0.1.1",
			current: "v0.1.0",
			latest:  "v0.1.1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "0.1.0-alpha vs v0.1.0",
			current: "0.1.0-alpha",
			latest:  "v0.1.0",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Invalid current version",
			current: "invalid",
			latest:  "v0.1.0",
			want:    false,
			wantErr: true,
		},
		{
			name:    "Invalid latest version",
			current: "v0.1.0",
			latest:  "invalid",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.current, tt.latest)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("CompareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name: "Simple version",
			version: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "1.2.3",
		},
		{
			name: "Version with prerelease",
			version: &Version{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: "alpha",
			},
			want: "0.1.0-alpha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
