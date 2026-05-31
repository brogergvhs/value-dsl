package semantictokens

import (
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
)

func mustAnalyzeText(t *testing.T, text string) coreanalysis.Result {
	t.Helper()

	result, err := coreanalysis.Run(text, coreanalysis.Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	return result
}
