package changesummary_test

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest/changesummary"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestWhereOpEquals(t *testing.T) {
	empty := &changesummary.ChangeSummary{}
	t.Run("empty match", func(t *testing.T) {
		matching := empty.WhereOpEquals(apitype.OpCreate)
		assert.Equal(t, empty, matching)
	})

	t.Run("filter out all", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}
		matching := summary.WhereOpEquals(apitype.OpDelete)
		assert.Equal(t, empty, matching)
	})

	t.Run("match", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}
		matching := summary.WhereOpEquals(apitype.OpCreate)
		assert.Equal(t, summary, matching)
	})

	t.Run("multiple matches", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate:            3,
			apitype.OpDelete:            2,
			apitype.OpCreateReplacement: 1,
		}
		matching := summary.WhereOpEquals(apitype.OpCreate, apitype.OpDelete)
		expected := &changesummary.ChangeSummary{
			apitype.OpCreate: 3,
			apitype.OpDelete: 2,
		}
		assert.Equal(t, expected, matching)
	})
}
func TestWhereOpNotEquals(t *testing.T) {
	empty := &changesummary.ChangeSummary{}
	t.Run("empty match", func(t *testing.T) {
		matching := empty.WhereOpNotEquals(apitype.OpCreate)
		assert.Equal(t, empty, matching)
	})

	t.Run("filter out all", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}
		matching := summary.WhereOpNotEquals(apitype.OpCreate)
		assert.Equal(t, empty, matching)
	})

	t.Run("exclude none", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}
		matching := summary.WhereOpNotEquals(apitype.OpDelete)
		assert.Equal(t, summary, matching)
	})

	t.Run("multiple matches", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpSame:              4,
			apitype.OpCreate:            3,
			apitype.OpDelete:            2,
			apitype.OpCreateReplacement: 1,
		}
		matching := summary.WhereOpNotEquals(apitype.OpCreate, apitype.OpDelete)
		expected := &changesummary.ChangeSummary{
			apitype.OpSame:              4,
			apitype.OpCreateReplacement: 1,
		}
		assert.Equal(t, expected, matching)
	})
}

func TestFromStringIntMap(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		summary := changesummary.FromStringIntMap(map[string]int{})
		assert.Equal(t, changesummary.ChangeSummary{}, summary)
	})

	t.Run("single", func(t *testing.T) {
		summary := changesummary.FromStringIntMap(map[string]int{
			"create": 1,
		})
		assert.Equal(t, changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}, summary)
	})

	t.Run("multiple", func(t *testing.T) {
		summary := changesummary.FromStringIntMap(map[string]int{
			"create": 3,
			"delete": 2,
		})
		assert.Equal(t, changesummary.ChangeSummary{
			apitype.OpCreate: 3,
			apitype.OpDelete: 2,
		}, summary)
	})
}

func TestChangeSummary_String(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var summary *changesummary.ChangeSummary
		assert.Equal(t, "", summary.String())
	})

	t.Run("empty", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{}
		assert.Equal(t, "", summary.String())
	})

	t.Run("single", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 1,
		}
		assert.Equal(t, "1 create", summary.String())
	})

	t.Run("multiple", func(t *testing.T) {
		summary := &changesummary.ChangeSummary{
			apitype.OpCreate: 3,
			apitype.OpDelete: 1,
		}
		assert.Equal(t, "3 creates, 1 delete", summary.String())
	})
}
