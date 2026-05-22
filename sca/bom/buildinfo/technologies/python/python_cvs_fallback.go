package python

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type pinnedRequirement struct {
	Name    string
	Version string
}

var pinnedRequirementRegex = regexp.MustCompile(`^\s*([A-Za-z0-9][A-Za-z0-9._-]*)(?:\[[^\]]*\])?\s*==\s*([^\s;]+)`)

func parseRequirementsTxtPins(reqPath string) ([]pinnedRequirement, error) {
	data, err := os.ReadFile(reqPath)
	if err != nil {
		return nil, fmt.Errorf("reading requirements file %s: %w", reqPath, err)
	}
	var pins []pinnedRequirement
	for _, raw := range strings.Split(string(data), "\n") {
		// pip's --hash uses `=` not `#`, so the first `#` is always a comment.
		if i := strings.Index(raw, "#"); i >= 0 {
			raw = raw[:i]
		}
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		m := pinnedRequirementRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pins = append(pins, pinnedRequirement{
			Name:    normalizePyPIName(m[1]),
			Version: m[2],
		})
	}
	return pins, nil
}

var pypiNameNormalizeRegex = regexp.MustCompile(`[-_.]+`)

func normalizePyPIName(name string) string {
	return strings.ToLower(pypiNameNormalizeRegex.ReplaceAllString(name, "-"))
}

func formatCvsBlockedRequirementsMessage(reqFile string, pins []pinnedRequirement) string {
	var b strings.Builder
	b.WriteString("Curation audit cannot complete: Artifactory CVS (Compliant Version Selection) is filtering blocked package versions out of the PyPI simple-index response — including the response served by the curation audit pass-through endpoint that `jf ca` relies on to enumerate dependencies. ")
	b.WriteString("Pip therefore reports the pinned version as missing instead of letting the curation block be detected.\n\n")
	if len(pins) > 0 {
		fmt.Fprintf(&b, "Pinned requirement(s) read from %s:\n", reqFile)
		for _, p := range pins {
			fmt.Fprintf(&b, "  - %s==%s\n", p.Name, p.Version)
		}
		b.WriteString("\n")
	}
	b.WriteString("To unblock the audit, disable CVS on the curated PyPI repository. Curation policies still apply at the file/download endpoint, so packages remain blocked from download — only the audit visibility is restored.\n")
	return b.String()
}

func isCvsVersionFilteredOutput(output string) bool {
	return strings.Contains(output, "No matching distribution found") ||
		strings.Contains(output, "Could not find a version that satisfies the requirement")
}
