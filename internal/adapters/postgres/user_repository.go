package postgres

import (
	"context"
	"errors"
	"time"

	domain "crud-api/internal/domain/user"
	"crud-api/internal/ports"

	"gorm.io/gorm"
)

type userModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:text;not null"`
	Email     string    `gorm:"type:text;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (userModel) TableName() string {
	return "users"
}

func Models() []any {
	return []any{&userModel{}}
}

func fromDomain(u *domain.User) *userModel {
	return &userModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (m *userModel) toDomain() *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	m := fromDomain(u)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&userModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []userModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	items := make([]domain.User, len(models))
	for i, m := range models {
		items[i] = *m.toDomain()
	}
	return items, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	m := fromDomain(u)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&userModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&userModel{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
