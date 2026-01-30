package core

import (
	"testing"
)

func TestConfigureAllowedTools_DefaultWhenEmpty(t *testing.T) {
	ConfigureAllowedTools(nil)

	got := SupportedTools()
	want := []string{"opencode", "amp"}

	if len(got) != len(want) {
		t.Fatalf("SupportedTools() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SupportedTools()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConfigureAllowedTools_CustomToolsOverrideDefaults(t *testing.T) {
	defer ConfigureAllowedTools(nil)

	ConfigureAllowedTools([]string{"amp", "custom-agent"})

	got := SupportedTools()
	want := []string{"amp", "custom-agent"}

	if len(got) != len(want) {
		t.Fatalf("SupportedTools() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SupportedTools()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConfigureAllowedTools_EmptySliceResetsToDefault(t *testing.T) {
	defer ConfigureAllowedTools(nil)

	ConfigureAllowedTools([]string{"custom"})
	ConfigureAllowedTools([]string{})

	got := SupportedTools()
	want := []string{"opencode", "amp"}

	if len(got) != len(want) {
		t.Fatalf("SupportedTools() after reset = %v, want %v", got, want)
	}
}

func TestConfigureAllowedTools_RemovesDuplicates(t *testing.T) {
	defer ConfigureAllowedTools(nil)

	ConfigureAllowedTools([]string{"amp", "amp", "opencode", "amp"})

	got := SupportedTools()
	want := []string{"amp", "opencode"}

	if len(got) != len(want) {
		t.Fatalf("SupportedTools() = %v, want %v", got, want)
	}
}

func TestIsSupportedTool_DefaultTools(t *testing.T) {
	ConfigureAllowedTools(nil)

	if !IsSupportedTool("opencode") {
		t.Error("IsSupportedTool(opencode) = false, want true")
	}
	if !IsSupportedTool("amp") {
		t.Error("IsSupportedTool(amp) = false, want true")
	}
	if IsSupportedTool("nonexistent") {
		t.Error("IsSupportedTool(nonexistent) = true, want false")
	}
}

func TestIsSupportedTool_CustomTools(t *testing.T) {
	defer ConfigureAllowedTools(nil)

	ConfigureAllowedTools([]string{"custom-agent"})

	if !IsSupportedTool("custom-agent") {
		t.Error("IsSupportedTool(custom-agent) = false, want true")
	}
	if IsSupportedTool("amp") {
		t.Error("IsSupportedTool(amp) = true, want false (not in custom list)")
	}
	if IsSupportedTool("opencode") {
		t.Error("IsSupportedTool(opencode) = true, want false (not in custom list)")
	}
}

func TestConfigureAllowedToolsFromEnv(t *testing.T) {
	defer ConfigureAllowedTools(nil)

	tests := []struct {
		name   string
		envVal string
		want   []string
	}{
		{
			name:   "empty env uses defaults",
			envVal: "",
			want:   []string{"opencode", "amp"},
		},
		{
			name:   "comma separated",
			envVal: "amp,custom,opencode",
			want:   []string{"amp", "custom", "opencode"},
		},
		{
			name:   "space separated",
			envVal: "tool1 tool2 tool3",
			want:   []string{"tool1", "tool2", "tool3"},
		},
		{
			name:   "mixed separators",
			envVal: "amp, custom\topencode",
			want:   []string{"amp", "custom", "opencode"},
		},
		{
			name:   "whitespace only resets to defaults",
			envVal: "   ",
			want:   []string{"opencode", "amp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigureAllowedTools(nil)

			t.Setenv("SOLO_TOOLS", tt.envVal)
			ConfigureAllowedToolsFromEnv()

			got := SupportedTools()
			if len(got) != len(tt.want) {
				t.Fatalf("SupportedTools() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("SupportedTools()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSupportedTools_ReturnsCopy(t *testing.T) {
	ConfigureAllowedTools(nil)

	got := SupportedTools()
	got[0] = "modified"

	original := SupportedTools()
	if original[0] == "modified" {
		t.Error("SupportedTools() returned a reference, not a copy")
	}
}
