package grpclog_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/grpclog"
	"github.com/stretchr/testify/assert"
)

func TestGrpcLog(t *testing.T) {
	log, err := grpclog.LoadLog(filepath.Join("testdata", "aws_bucket_grpc.json"))
	assert.NoError(t, err)
	assert.Equal(t, 71, len(log.Entries))

	t.Run("where filter", func(t *testing.T) {
		creates := log.WhereMethod(grpclog.Create)
		assert.Equal(t, 2, len(creates))
	})

	t.Run("typed creates", func(t *testing.T) {
		creates, err := log.Creates()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(creates))
		// Check deserialized value
		assert.Equal(t, "urn:pulumi:p-it-antons-mac-bucket-9f59db4a::test::aws:s3/bucket:Bucket::tested-resource", creates[0].Request.Urn)
	})

	t.Run("serialize roundtrip", func(t *testing.T) {
		bytes, err := log.Marshal()
		assert.NoError(t, err)
		unmarshalled, err := grpclog.ParseLog(bytes)
		assert.NoError(t, err)
		assert.Equal(t, log, unmarshalled)
	})
}
