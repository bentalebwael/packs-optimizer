package order_calculations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

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

func TestSave(t *testing.T) {
	t.Run("successful save", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		packResult := []PackResult{
			{Size: 250, Quantity: 2},
			{Size: 500, Quantity: 1},
		}
		_, err := json.Marshal(packResult)
		require.NoError(t, err)

		calc := &OrderCalculation{
			OrderQuantity:   1250,
			Result:          packResult,
			TotalItems:      1250,
			TotalPacks:      3,
			ConfigurationID: 1,
			Timestamp:       time.Now(),
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Use sqlmock.AnyArg() for the JSON result to avoid type comparison issues
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "order_calculations" ("order_quantity","result","total_items","total_packs","configuration_id","timestamp") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id","timestamp"`)).
			WithArgs(calc.OrderQuantity, sqlmock.AnyArg(), calc.TotalItems, calc.TotalPacks, calc.ConfigurationID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "timestamp"}).AddRow(1, time.Now()))

		// Expect the COMMIT
		mock.ExpectCommit()

		// Execute
		err = repo.Save(ctx, calc)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, uint(1), calc.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()

		packResult := []PackResult{
			{Size: 250, Quantity: 2},
			{Size: 500, Quantity: 1},
		}
		_, err := json.Marshal(packResult)
		require.NoError(t, err)

		calc := &OrderCalculation{
			OrderQuantity:   1250,
			Result:          packResult,
			TotalItems:      1250,
			TotalPacks:      3,
			ConfigurationID: 1,
			Timestamp:       time.Now(),
		}

		// Expect the BEGIN transaction
		mock.ExpectBegin()

		// Use sqlmock.AnyArg() for the JSON result to avoid type comparison issues
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "order_calculations" ("order_quantity","result","total_items","total_packs","configuration_id","timestamp") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id","timestamp"`)).
			WithArgs(calc.OrderQuantity, sqlmock.AnyArg(), calc.TotalItems, calc.TotalPacks, calc.ConfigurationID, sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		// Expect the ROLLBACK
		mock.ExpectRollback()

		// Execute
		err = repo.Save(ctx, calc)

		// Assert
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetByID(t *testing.T) {
	t.Run("get existing calculation", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(1)

		packResult := []PackResult{
			{Size: 250, Quantity: 2},
			{Size: 500, Quantity: 1},
		}
		resultJSON, err := json.Marshal(packResult)
		require.NoError(t, err)

		timestamp := time.Now()

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE "order_calculations"."id" = $1 ORDER BY "order_calculations"."id" LIMIT $2`)).
			WithArgs(id, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_quantity", "result", "total_items", "total_packs", "configuration_id", "timestamp"}).
				AddRow(1, 1250, resultJSON, 1250, 3, 1, timestamp))

		// Expect SELECT query for the configuration (preload)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{250, 500, 1000}, "test-signature", true))

		// Execute
		result, err := repo.GetByID(ctx, id)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, 1250, result.OrderQuantity)
		assert.Equal(t, packResult, result.Result)
		assert.Equal(t, 1250, result.TotalItems)
		assert.Equal(t, 3, result.TotalPacks)
		assert.Equal(t, uint(1), result.ConfigurationID)
		assert.Equal(t, timestamp.UTC(), result.Timestamp.UTC())
		assert.NotNil(t, result.Configuration)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		id := uint(999)

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE "order_calculations"."id" = $1 ORDER BY "order_calculations"."id" LIMIT $2`)).
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

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE "order_calculations"."id" = $1 ORDER BY "order_calculations"."id" LIMIT $2`)).
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

func TestGetByConfigurationIDAndOrderQuantity(t *testing.T) {
	t.Run("get existing calculation", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		orderQuantity := 1250
		configID := uint(1)

		packResult := []PackResult{
			{Size: 250, Quantity: 2},
			{Size: 500, Quantity: 1},
		}
		resultJSON, err := json.Marshal(packResult)
		require.NoError(t, err)

		timestamp := time.Now()

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE order_quantity = $1 AND configuration_id = $2 ORDER BY "order_calculations"."id" LIMIT $3`)).
			WithArgs(orderQuantity, configID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_quantity", "result", "total_items", "total_packs", "configuration_id", "timestamp"}).
				AddRow(1, orderQuantity, resultJSON, 1250, 3, configID, timestamp))

		// Expect SELECT query for the configuration (preload)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1`)).
			WithArgs(configID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{250, 500, 1000}, "test-signature", true))

		// Execute
		result, err := repo.GetByConfigurationIDAndOrderQuantity(ctx, orderQuantity, configID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, orderQuantity, result.OrderQuantity)
		assert.Equal(t, configID, result.ConfigurationID)
		assert.Equal(t, packResult, result.Result)
		assert.NotNil(t, result.Configuration)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		orderQuantity := 1250
		configID := uint(999)

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE order_quantity = $1 AND configuration_id = $2 ORDER BY "order_calculations"."id" LIMIT $3`)).
			WithArgs(orderQuantity, configID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Execute
		result, err := repo.GetByConfigurationIDAndOrderQuantity(ctx, orderQuantity, configID)

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
		orderQuantity := 1250
		configID := uint(1)

		// Update to match GORM's parameterized LIMIT query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" WHERE order_quantity = $1 AND configuration_id = $2 ORDER BY "order_calculations"."id" LIMIT $3`)).
			WithArgs(orderQuantity, configID, 1).
			WillReturnError(errors.New("database error"))

		// Execute
		result, err := repo.GetByConfigurationIDAndOrderQuantity(ctx, orderQuantity, configID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestList(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		offset := 0
		limit := 10

		timestamp := time.Now()
		packResult1 := []PackResult{{Size: 250, Quantity: 2}, {Size: 500, Quantity: 1}}
		packResult2 := []PackResult{{Size: 500, Quantity: 1}, {Size: 1000, Quantity: 1}}
		resultJSON1, _ := json.Marshal(packResult1)
		resultJSON2, _ := json.Marshal(packResult2)

		// Expect SELECT query for calculations
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" ORDER BY timestamp DESC LIMIT $1`)).
			WithArgs(limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_quantity", "result", "total_items", "total_packs", "configuration_id", "timestamp"}).
				AddRow(1, 1250, resultJSON1, 1250, 3, 1, timestamp).
				AddRow(2, 1500, resultJSON2, 1500, 2, 1, timestamp))

		// Expect SELECT query for the configuration (preload)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pack_configurations" WHERE "pack_configurations"."id" = $1`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "pack_sizes", "signature", "active"}).
				AddRow(1, pq.Int64Array{250, 500, 1000}, "test-signature", true))

		// Execute
		results, err := repo.List(ctx, offset, limit)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, uint(1), results[0].ID)
		assert.Equal(t, packResult1, results[0].Result)
		assert.Equal(t, uint(2), results[1].ID)
		assert.Equal(t, packResult2, results[1].Result)
		assert.NotNil(t, results[0].Configuration)
		assert.NotNil(t, results[1].Configuration)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		// Setup
		db, mock, sqlDB := setupTestDB(t)
		defer sqlDB.Close()

		repo := NewRepository(db)
		ctx := context.Background()
		offset := 0
		limit := 10

		// Expect SELECT query returning empty result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" ORDER BY timestamp DESC LIMIT $1`)).
			WithArgs(limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "order_quantity", "result", "total_items", "total_packs", "configuration_id", "timestamp"}))

		// Execute
		results, err := repo.List(ctx, offset, limit)

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
		offset := 0
		limit := 10

		// Expect SELECT query with error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_calculations" ORDER BY timestamp DESC LIMIT $1`)).
			WithArgs(limit).
			WillReturnError(errors.New("database error"))

		// Execute
		results, err := repo.List(ctx, offset, limit)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "order_calculations" WHERE "order_calculations"."id" = $1`)).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "order_calculations" WHERE "order_calculations"."id" = $1`)).
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
