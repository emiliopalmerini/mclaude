package ports

import (
	"context"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
)

type PricingRepository interface {
	Create(ctx context.Context, pricing *domain.ModelPricing) error
	GetByID(ctx context.Context, id string) (*domain.ModelPricing, error)
	GetDefault(ctx context.Context) (*domain.ModelPricing, error)
	List(ctx context.Context) ([]*domain.ModelPricing, error)
	Update(ctx context.Context, pricing *domain.ModelPricing) error
	SetDefault(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
