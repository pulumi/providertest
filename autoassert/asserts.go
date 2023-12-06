package autoassert

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// UpHasNoChanges asserts that the given UpResult has no changes - only "same" operations allowed.
func UpHasNoChanges(t *testing.T, up auto.UpResult) {
	t.Helper()

	unexpectedOps := getUnexpectedOps(*up.Summary.ResourceChanges, apitype.OpSame)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

// UpHasNoDeletes asserts that the given UpResult has no deletes - only "same", "create", "update", "refresh", and
func UpHasNoDeletes(t *testing.T, up auto.UpResult) {
	t.Helper()

	unexpectedOps := findMatchingOps(*up.Summary.ResourceChanges, apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

func PreviewHasNoChanges(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := opMapToStringMap(preview.ChangeSummary)
	unexpectedOps := getUnexpectedOps(convertedMap, apitype.OpSame)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

func PreviewHasNoDeletes(t *testing.T, preview auto.PreviewResult) {
	t.Helper()

	convertedMap := opMapToStringMap(preview.ChangeSummary)
	unexpectedOps := findMatchingOps(convertedMap, apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReplace)

	if len(unexpectedOps) > 0 {
		t.Errorf("expected no changes, got %s", strings.Join(unexpectedOps, ", "))
	}
}

func getUnexpectedOps(changeSummary map[string]int, expectedOpTypes ...apitype.OpType) []string {
	// Sort the op types so that the output is deterministic.
	orderedOpTypes := []string{}
	for opType := range changeSummary {
		orderedOpTypes = append(orderedOpTypes, opType)
	}
	sort.SliceStable(orderedOpTypes, func(i, j int) bool {
		return orderedOpTypes[i] < orderedOpTypes[j]
	})

	var unexpectedOps []string
	for _, opType := range orderedOpTypes {
		opCount := changeSummary[opType]
		if opCount == 0 {
			continue
		}
		isExpected := false
		for _, expectedOpType := range expectedOpTypes {
			if opType == string(expectedOpType) {
				isExpected = true
				break
			}
		}
		if isExpected {
			continue
		}
		unexpectedOps = append(unexpectedOps, formatOpCount(opType, opCount))
	}
	return unexpectedOps
}

func findMatchingOps(changeSummary map[string]int, opTypes ...apitype.OpType) []string {
	// Sort the op types so that the output is deterministic.
	orderedOpTypes := []string{}
	for opType := range changeSummary {
		orderedOpTypes = append(orderedOpTypes, opType)
	}
	sort.SliceStable(orderedOpTypes, func(i, j int) bool {
		return orderedOpTypes[i] < orderedOpTypes[j]
	})

	var matchingOps []string
	for _, opType := range orderedOpTypes {
		opCount := changeSummary[opType]
		if opCount == 0 {
			continue
		}
		for _, expectedOpTypes := range opTypes {
			for _, expectedOpType := range expectedOpTypes {
				if opType == string(expectedOpType) {
					matchingOps = append(matchingOps, formatOpCount(opType, opCount))
				}
			}
		}
	}
	return matchingOps
}

func formatOpCount(opType string, count int) string {
	pluralSuffix := ""
	if count > 1 {
		pluralSuffix = "s"
	}
	return fmt.Sprintf("%d %s%s", count, opType, pluralSuffix)
}

func opMapToStringMap(opMap map[apitype.OpType]int) map[string]int {
	strMap := map[string]int{}
	for opType, count := range opMap {
		strMap[string(opType)] = count
	}
	return strMap
}
