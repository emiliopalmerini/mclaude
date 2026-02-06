package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type PricingRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPricingRepository(db *sql.DB) *PricingRepository {
	return &PricingRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PricingRepository) Create(ctx context.Context, pricing *domain.ModelPricing) error {
	return r.queries.CreateModelPricing(ctx, sqlc.CreateModelPricingParams{
		ID:                          pricing.ID,
		DisplayName:                 pricing.DisplayName,
		InputPerMillion:             pricing.InputPerMillion,
		OutputPerMillion:            pricing.OutputPerMillion,
		CacheReadPerMillion:         util.NullFloat64(pricing.CacheReadPerMillion),
		CacheWritePerMillion:        util.NullFloat64(pricing.CacheWritePerMillion),
		LongContextInputPerMillion:  util.NullFloat64(pricing.LongContextInputPerMillion),
		LongContextOutputPerMillion: util.NullFloat64(pricing.LongContextOutputPerMillion),
		LongContextThreshold:        util.NullInt64(pricing.LongContextThreshold),
		IsDefault:                   util.BoolToInt64(pricing.IsDefault),
		CreatedAt:                   pricing.CreatedAt.Format(time.RFC3339),
	})
}

func (r *PricingRepository) GetByID(ctx context.Context, id string) (*domain.ModelPricing, error) {
	row, err := r.queries.GetModelPricingByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get model pricing: %w", err)
	}
	return pricingFromRow(row), nil
}

func (r *PricingRepository) GetDefault(ctx context.Context) (*domain.ModelPricing, error) {
	row, err := r.queries.GetDefaultModelPricing(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default model pricing: %w", err)
	}
	return pricingFromRow(row), nil
}

func (r *PricingRepository) List(ctx context.Context) ([]*domain.ModelPricing, error) {
	rows, err := r.queries.ListModelPricing(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list model pricing: %w", err)
	}

	pricing := make([]*domain.ModelPricing, len(rows))
	for i, row := range rows {
		pricing[i] = pricingFromRow(row)
	}
	return pricing, nil
}

func (r *PricingRepository) Update(ctx context.Context, pricing *domain.ModelPricing) error {
	return r.queries.UpdateModelPricing(ctx, sqlc.UpdateModelPricingParams{
		DisplayName:                 pricing.DisplayName,
		InputPerMillion:             pricing.InputPerMillion,
		OutputPerMillion:            pricing.OutputPerMillion,
		CacheReadPerMillion:         util.NullFloat64(pricing.CacheReadPerMillion),
		CacheWritePerMillion:        util.NullFloat64(pricing.CacheWritePerMillion),
		LongContextInputPerMillion:  util.NullFloat64(pricing.LongContextInputPerMillion),
		LongContextOutputPerMillion: util.NullFloat64(pricing.LongContextOutputPerMillion),
		LongContextThreshold:        util.NullInt64(pricing.LongContextThreshold),
		IsDefault:                   util.BoolToInt64(pricing.IsDefault),
		ID:                          pricing.ID,
	})
}

func (r *PricingRepository) SetDefault(ctx context.Context, id string) error {
	return r.queries.SetDefaultModelPricing(ctx, id)
}

func (r *PricingRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteModelPricing(ctx, id)
}

func pricingFromRow(row sqlc.ModelPricing) *domain.ModelPricing {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)

	var cacheReadPerMillion, cacheWritePerMillion *float64
	if row.CacheReadPerMillion.Valid {
		cacheReadPerMillion = &row.CacheReadPerMillion.Float64
	}
	if row.CacheWritePerMillion.Valid {
		cacheWritePerMillion = &row.CacheWritePerMillion.Float64
	}

	var longContextInputPerMillion, longContextOutputPerMillion *float64
	var longContextThreshold *int64
	if row.LongContextInputPerMillion.Valid {
		longContextInputPerMillion = &row.LongContextInputPerMillion.Float64
	}
	if row.LongContextOutputPerMillion.Valid {
		longContextOutputPerMillion = &row.LongContextOutputPerMillion.Float64
	}
	if row.LongContextThreshold.Valid {
		longContextThreshold = &row.LongContextThreshold.Int64
	}

	return &domain.ModelPricing{
		ID:                          row.ID,
		DisplayName:                 row.DisplayName,
		InputPerMillion:             row.InputPerMillion,
		OutputPerMillion:            row.OutputPerMillion,
		CacheReadPerMillion:         cacheReadPerMillion,
		CacheWritePerMillion:        cacheWritePerMillion,
		LongContextInputPerMillion:  longContextInputPerMillion,
		LongContextOutputPerMillion: longContextOutputPerMillion,
		LongContextThreshold:        longContextThreshold,
		IsDefault:                   row.IsDefault == 1,
		CreatedAt:                   createdAt,
	}
}
