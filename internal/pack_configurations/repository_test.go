package pack_configurations

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	// Create a new SQL mock
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create sqlmock")

	// Create GORM dialector using the mock DB
	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	})

	// Open GORM DB with the dialector
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err, "Failed to open GORM DB with mock")

	return db, mock, sqlDB
}

// TestCreate tests the Create method
func TestCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		config := &PackConfiguration{
			Signature: "test-signature",
			PackSizes: pq.Int64Array{1, 2, 3},
			Active:    false,
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect INSERT query with returning ID
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pack_configurations" ("pack_sizes","signature","active") VALUES ($1,$2,$3) RETURNING "id"`)).
			WithArgs(config.PackSizes, config.Signature, config.Active).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect the COMMIT
		mock.ExpectCommit()

		// Execute
		result, err := repo.Create(ctx, config)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		config := &PackConfiguration{
			Signature: "test-signature",
			PackSizes: pq.Int64Array{1, 2, 3},
			Active:    false,
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect INSERT query with an error
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pack_configurations" ("pack_sizes","signature","active") VALUES ($1,$2,$3) RETURNING "id"`)).
			WithArgs(config.PackSizes, config.Signature, config.Active).
			WillReturnError(errors.New("database error"))

		// Expect the ROLLBACK
		mock.ExpectRollback()

		// Execute
		result, err := repo.Create(ctx, config)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, config, result) // The input struct is returned on error
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetByID tests the GetByID method
func TestGetByID(t *testing.T) {
	t.Run("get existing configuration", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)
		limit := uint(1)

		// Expect SELECT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(id, limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{1, 2, 3}, "test-signature", false))

		// Execute
		result, err := repo.GetByID(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, pq.Int64Array{1, 2, 3}, result.PackSizes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(999)

		// Expect SELECT query that returns not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(id, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Execute
		result, err := repo.GetByID(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect SELECT query that returns an error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(id, 1).
			WillReturnError(errors.New("database error"))

		// Execute
		result, err := repo.GetByID(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetBySignature tests the GetBySignature method
func TestGetBySignature(t *testing.T) {
	t.Run("get existing configuration", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		signature := "test-signature"

		// Expect SELECT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE signature = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(signature, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{1, 2, 3}, signature, false))

		// Execute
		result, err := repo.GetBySignature(ctx, signature)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, signature, result.Signature)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, pq.Int64Array{1, 2, 3}, result.PackSizes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		signature := "non-existent-signature"

		// Expect SELECT query that returns not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE signature = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(signature, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Execute
		result, err := repo.GetBySignature(ctx, signature)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		signature := "test-signature"

		// Expect SELECT query that returns an error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE signature = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(signature, 1).
			WillReturnError(errors.New("database error"))

		// Execute
		result, err := repo.GetBySignature(ctx, signature)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetActive tests the GetActive method
func TestGetActive(t *testing.T) {
	t.Run("get active configuration", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query for active config
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE active = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(true, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{1, 2, 3}, "active-signature", true))

		// Execute
		result, err := repo.GetActive(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Active)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, pq.Int64Array{1, 2, 3}, result.PackSizes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no active configuration", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query that returns no active config
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE active = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(true, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Execute
		result, err := repo.GetActive(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query that returns an error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE active = $1 ORDER BY "pack_configurations"."id" LIMIT $2`)).
			WithArgs(true, 1).
			WillReturnError(errors.New("database error"))

		// Execute
		result, err := repo.GetActive(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdate tests the Update method
func TestUpdate(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		config := &PackConfiguration{
			ID:        1,
			PackSizes: pq.Int64Array{4, 5, 6},
			Signature: "updated-signature",
			Active:    true,
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect UPDATE query
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "pack_sizes"=$1,"signature"=$2,"active"=$3 WHERE "id" = $4`)).
			WithArgs(config.PackSizes, config.Signature, config.Active, config.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect the COMMIT
		mock.ExpectCommit()

		// Execute
		err := repo.Update(ctx, config)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		config := &PackConfiguration{
			ID:        1,
			PackSizes: pq.Int64Array{4, 5, 6},
			Signature: "updated-signature",
			Active:    true,
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect UPDATE query with error
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "pack_sizes"=$1,"signature"=$2,"active"=$3 WHERE "id" = $4`)).
			WithArgs(config.PackSizes, config.Signature, config.Active, config.ID).
			WillReturnError(errors.New("database error"))

		// Expect the ROLLBACK
		mock.ExpectRollback()

		// Execute
		err := repo.Update(ctx, config)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDelete tests the Delete method
func TestDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect DELETE query
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "pack_configurations" WHERE "pack_configurations"."id" = $1`)).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect the COMMIT
		mock.ExpectCommit()

		// Execute
		err := repo.Delete(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Expect DELETE query with error
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "pack_configurations" WHERE "pack_configurations"."id" = $1`)).
			WithArgs(id).
			WillReturnError(errors.New("database error"))

		// Expect the ROLLBACK
		mock.ExpectRollback()

		// Execute
		err := repo.Delete(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestList tests the List method
func TestList(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{1, 2, 3}, "signature-1", true).
				AddRow(2, pq.Int64Array{4, 5, 6}, "signature-2", false))

		// Execute
		results, err := repo.List(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, uint(1), results[0].ID)
		assert.Equal(t, pq.Int64Array{1, 2, 3}, results[0].PackSizes)
		assert.Equal(t, uint(2), results[1].ID)
		assert.Equal(t, pq.Int64Array{4, 5, 6}, results[1].PackSizes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query returning empty result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}))

		// Execute
		results, err := repo.List(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		// Expect SELECT query with error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations"`)).
			WillReturnError(errors.New("database error"))

		// Execute
		results, err := repo.List(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestSetActive tests the SetActive method
func TestSetActive(t *testing.T) {
	t.Run("set configuration as active", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect begin transaction
		mock.ExpectBegin()

		// Expect update to deactivate current active
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE active = $2`)).
			WithArgs(false, true).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect update to activate new one
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE id = $2`)).
			WithArgs(true, id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect commit
		mock.ExpectCommit()

		// Execute
		err := repo.SetActive(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no active configuration to deactivate", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect begin transaction
		mock.ExpectBegin()

		// Expect update to deactivate (returns 0 rows affected)
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE active = $2`)).
			WithArgs(false, true).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Expect update to activate
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE id = $2`)).
			WithArgs(true, id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect commit
		mock.ExpectCommit()

		// Execute
		err := repo.SetActive(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on first update", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect begin transaction
		mock.ExpectBegin()

		// Expect update with error
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE active = $2`)).
			WithArgs(false, true).
			WillReturnError(errors.New("database error"))

		// Expect rollback
		mock.ExpectRollback()

		// Execute
		err := repo.SetActive(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on second update", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		// Expect begin transaction
		mock.ExpectBegin()

		// Expect first update success
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE active = $2`)).
			WithArgs(false, true).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect second update with error
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE id = $2`)).
			WithArgs(true, id).
			WillReturnError(errors.New("database error"))

		// Expect rollback
		mock.ExpectRollback()

		// Execute
		err := repo.SetActive(ctx, id)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set non-existent configuration as active", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(999) // Non-existent ID

		// Expect begin transaction
		mock.ExpectBegin()

		// Expect update to deactivate
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE active = $2`)).
			WithArgs(false, true).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Expect update to activate non-existent returns 0 rows
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "pack_configurations" SET "active"=$1 WHERE id = $2`)).
			WithArgs(true, id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Expect commit
		mock.ExpectCommit()

		// Execute
		err := repo.SetActive(ctx, id)

		// Assert
		assert.NoError(t, err) // GORM doesn't return error when updated rows is 0
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
