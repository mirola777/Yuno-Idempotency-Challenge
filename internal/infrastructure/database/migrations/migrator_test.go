package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationRecordTableName(t *testing.T) {
	record := MigrationRecord{}
	assert.Equal(t, "schema_migrations", record.TableName())
}

func TestRegisterAddsMigrationToRegistry(t *testing.T) {
	original := registry
	defer func() { registry = original }()

	registry = nil

	Register(Migration{ID: "test_001"})
	assert.Len(t, registry, 1)
	assert.Equal(t, "test_001", registry[0].ID)

	Register(Migration{ID: "test_002"})
	assert.Len(t, registry, 2)
	assert.Equal(t, "test_002", registry[1].ID)
}

func TestRegistryPreservesOrdering(t *testing.T) {
	original := registry
	defer func() { registry = original }()

	registry = nil

	ids := []string{"alpha", "beta", "gamma", "delta"}
	for _, id := range ids {
		Register(Migration{ID: id})
	}

	assert.Len(t, registry, len(ids))
	for i, id := range ids {
		assert.Equal(t, id, registry[i].ID)
	}
}
