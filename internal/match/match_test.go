package match

import "testing"

func TestMatches(t *testing.T) {
	cases := map[string]struct {
		path       string
		globs      []string
		extensions []string
		want       bool
	}{
		"extension match":          {"/z/a.txt", nil, []string{".txt"}, true},
		"extension without dot":    {"/z/a.txt", nil, []string{"txt"}, true},
		"extension case-insensitive": {"/z/A.TXT", nil, []string{".txt"}, true},
		"extension mismatch":       {"/z/a.log", nil, []string{".txt"}, false},
		"glob match":               {"/z/a.txt", []string{"*.txt"}, nil, true},
		"glob mismatch":            {"/z/a.log", []string{"*.txt"}, nil, false},
		"both apply":               {"/z/a.txt", []string{"a.*"}, []string{".txt"}, true},
		"glob fails short-circuits": {"/z/a.txt", []string{"b.*"}, []string{".txt"}, false},
		"ext fails short-circuits": {"/z/a.log", []string{"*.*"}, []string{".txt"}, false},
		"invalid glob ignored":     {"/z/a.txt", []string{"[bad"}, []string{".txt"}, false},
		"matches basename not path": {"/txt/a.log", []string{"*.txt"}, nil, false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := Matches(tc.path, tc.globs, tc.extensions); got != tc.want {
				t.Errorf("Matches() = %v, want %v", got, tc.want)
			}
		})
	}
}
