package git

import "testing"

func TestParseSubmoduleStatusLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
		ok   bool
	}{
		{name: "space prefix", line: " 0123456789abcdef apps/maestro (heads/main)", want: "apps/maestro", ok: true},
		{name: "dash prefix", line: "-0123456789abcdef packages/claude-mem (heads/main)", want: "packages/claude-mem", ok: true},
		{name: "plus prefix", line: "+0123456789abcdef submodules/hyperharness (remotes/origin/main)", want: "submodules/hyperharness", ok: true},
		{name: "conflict prefix", line: "U0123456789abcdef archive/OmniRoute (v1.0.0)", want: "archive/OmniRoute", ok: true},
		{name: "blank", line: "   ", want: "", ok: false},
		{name: "malformed", line: "abcdefonly", want: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseSubmoduleStatusLine(tc.line)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if got != tc.want {
				t.Fatalf("expected path %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseSubmoduleStatusOutputSortsPaths(t *testing.T) {
	output := "+0123 submodules/hyperharness (heads/main)\n 4567 apps/maestro (heads/main)\n-89ab packages/claude-mem (heads/main)\n"
	got := parseSubmoduleStatusOutput(output)
	want := []string{"apps/maestro", "packages/claude-mem", "submodules/hyperharness"}

	if len(got) != len(want) {
		t.Fatalf("expected %d paths, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected path[%d]=%q, got %q", i, want[i], got[i])
		}
	}
}

func TestFilepathBaseSupportsSlashAndBackslash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "apps/maestro", want: "maestro"},
		{input: `submodules\hyperharness`, want: "hyperharness"},
		{input: "", want: ""},
	}

	for _, tc := range tests {
		if got := filepathBase(tc.input); got != tc.want {
			t.Fatalf("filepathBase(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
