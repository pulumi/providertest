package asserthelpers

import (
	"fmt"
	"sort"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func FindUnexpectedOps(changeSummary map[string]int, expectedOpTypes ...apitype.OpType) []string {
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

func FindMatchingOps(changeSummary map[string]int, opTypes ...apitype.OpType) []string {
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

func OpMapToStringMap(opMap map[apitype.OpType]int) map[string]int {
	strMap := map[string]int{}
	for opType, count := range opMap {
		strMap[string(opType)] = count
	}
	return strMap
}
