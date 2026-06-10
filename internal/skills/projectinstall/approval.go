package projectinstall

import (
	"regexp"
	"strings"
)

var approvedEvidencePattern = regexp.MustCompile(`^dry-run:sha256:[0-9a-f]{64}$`)

func approvedHashFromEvidence(evidenceRef string) (string, bool) {
	if !approvedEvidencePattern.MatchString(evidenceRef) {
		hash, _ := strings.CutPrefix(evidenceRef, "dry-run:")
		return hash, false
	}
	return strings.TrimPrefix(evidenceRef, "dry-run:"), true
}
