package conv

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Olahraga", want: "olahraga"},
		{input: "Teknologi Terkini", want: "teknologi-terkini"},
		{input: "Tech & AI / 2026!", want: "tech-ai-2026"},
		{input: "  Multiple   Spaces---Here  ", want: "multiple-spaces-here"},
	}

	for _, tc := range tests {
		if got := GenerateSlug(tc.input); got != tc.want {
			t.Errorf("GenerateSlug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
