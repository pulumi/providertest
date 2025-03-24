package changesummary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

func (changeSummary *ChangeSummary) WhereOpNotEquals(opTypes ...apitype.OpType) *ChangeSummary {
	if changeSummary == nil {
		return nil
	}
	input := *changeSummary
	// Sort the op types so that the output is deterministic.
	orderedOpTypes := []apitype.OpType{}
	for opType := range input {
		orderedOpTypes = append(orderedOpTypes, opType)
	}
	sort.SliceStable(orderedOpTypes, func(i, j int) bool {
		return orderedOpTypes[i] < orderedOpTypes[j]
	})

	result := ChangeSummary{}
	for _, opType := range orderedOpTypes {
		opCount := input[opType]
		if opCount == 0 {
			continue
		}
		isExpected := false
		for _, expectedOpType := range opTypes {
			if opType == expectedOpType {
				isExpected = true
				break
			}
		}
		if isExpected {
			continue
		}
		result[opType] = opCount
	}
	return &result
}

func (changeSummary *ChangeSummary) WhereOpEquals(opTypes ...apitype.OpType) *ChangeSummary {
	if changeSummary == nil {
		return nil
	}
	input := *changeSummary
	// Sort the op types so that the output is deterministic.
	orderedOpTypes := []apitype.OpType{}
	for opType := range input {
		orderedOpTypes = append(orderedOpTypes, opType)
	}
	sort.SliceStable(orderedOpTypes, func(i, j int) bool {
		return orderedOpTypes[i] < orderedOpTypes[j]
	})

	matching := ChangeSummary{}
	for _, opType := range orderedOpTypes {
		opCount := input[opType]
		if opCount == 0 {
			continue
		}
		for _, expectedOpType := range opTypes {
			if opType == expectedOpType {
				matching[opType] = opCount
			}
		}
	}
	return &matching
}

func FromStringIntMap(input map[string]int) ChangeSummary {
	result := map[apitype.OpType]int{}
	for opType, count := range input {
		result[apitype.OpType(opType)] = count
	}
	return result
}

type ChangeSummary map[apitype.OpType]int

func (changeSummary *ChangeSummary) String() string {
	if changeSummary == nil {
		return ""
	}
	input := *changeSummary
	sortedOps := []apitype.OpType{}
	for opType := range input {
		sortedOps = append(sortedOps, opType)
	}
	sort.SliceStable(sortedOps, func(i, j int) bool {
		return sortedOps[i] < sortedOps[j]
	})
	var formatted []string
	for _, opType := range sortedOps {
		count := input[opType]
		pluralSuffix := ""
		if count > 1 {
			pluralSuffix = "s"
		}
		formatted = append(formatted, fmt.Sprintf("%d %s%s", count, opType, pluralSuffix))
	}
	return strings.Join(formatted, ", ")
}
