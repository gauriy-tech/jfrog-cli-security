package python

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRequirementsTxtPins(t *testing.T) {
	cases := []struct {
		name     string
		contents string
		want     []pinnedRequirement
	}{
		{
			name: "extracts pins, ignores comments/blank/options/non-pins, normalizes names",
			contents: "" +
				"# a comment\n" +
				"\n" +
				"-r other.txt\n" +
				"-e ./local\n" +
				"langchain>=1.2.15\n" +
				"requests[security]==2.34.2  # trailing\n" +
				`flask==3.0.0 ; python_version >= "3.9"` + "\n" +
				"Langchain_Core==1.4.0\n",
			want: []pinnedRequirement{
				{Name: "requests", Version: "2.34.2"},
				{Name: "flask", Version: "3.0.0"},
				{Name: "langchain-core", Version: "1.4.0"},
			},
		},
		{
			name:     "empty file yields no pins",
			contents: "",
			want:     nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "requirements.txt")
			require.NoError(t, os.WriteFile(path, []byte(tc.contents), 0o600))
			got, err := parseRequirementsTxtPins(path)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsCvsVersionFilteredOutput(t *testing.T) {
	cases := map[string]bool{
		"ERROR: No matching distribution found for deepagents==0.5.5":                                true,
		"ERROR: Could not find a version that satisfies the requirement langchain-core<2.0.0,>=1.3.2": true,
		"ERROR: 403 Forbidden":                                                                       false,
	}
	for output, want := range cases {
		t.Run(output, func(t *testing.T) {
			assert.Equal(t, want, isCvsVersionFilteredOutput(output))
		})
	}
}

func TestFormatCvsBlockedRequirementsMessage(t *testing.T) {
	msg := formatCvsBlockedRequirementsMessage("requirements.txt",
		[]pinnedRequirement{{Name: "deepagents", Version: "0.5.5"}})

	assert.Contains(t, msg, "deepagents==0.5.5")
	assert.Contains(t, msg, "requirements.txt")
	assert.Contains(t, msg, "CVS")
	assert.Contains(t, strings.ToLower(msg), "disable cvs")
}
