package decorator

import (
	"context"
	"time"

	"crud-api/internal/domain/user"
	"crud-api/internal/ports"
)

type cacheDecorator struct {
	base     ports.UserService
	cache    ports.UserCache
	cacheTTL time.Duration
}

func NewCacheDecorator(base ports.UserService, cache ports.UserCache, cacheTTL time.Duration) ports.UserService {
	return &cacheDecorator{
		base:     base,
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

func (d *cacheDecorator) Create(ctx context.Context, req ports.CreateUserInput) (*user.User, error) {
	return d.base.Create(ctx, req)
}

func (d *cacheDecorator) Get(ctx context.Context, id string) (*user.User, error) {
	if d.cache != nil {
		if u, err := d.cache.Get(ctx, id); err == nil && u != nil {
			return u, nil
		}
	}

	u, err := d.base.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if d.cache != nil {
		_ = d.cache.Set(ctx, u, d.cacheTTL)
	}

	return u, nil
}

func (d *cacheDecorator) List(ctx context.Context, page, pageSize int) ([]user.User, int64, error) {
	return d.base.List(ctx, page, pageSize)
}

func (d *cacheDecorator) Update(ctx context.Context, id string, req ports.UpdateUserInput) (*user.User, error) {
	u, err := d.base.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if d.cache != nil {
		_ = d.cache.Delete(ctx, id)
	}

	return u, nil
}

func (d *cacheDecorator) Delete(ctx context.Context, id string) error {
	if err := d.base.Delete(ctx, id); err != nil {
		return err
	}

	if d.cache != nil {
		_ = d.cache.Delete(ctx, id)
	}

	return nil
}
