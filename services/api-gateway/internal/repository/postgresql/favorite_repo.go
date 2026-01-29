package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/egrul-system/services/api-gateway/internal/graph/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// FavoriteRepository реализация для PostgreSQL
type FavoriteRepository struct {
	db     *sql.DB
	schema string
	logger *zap.Logger
}

// NewFavoriteRepository создает новый экземпляр FavoriteRepository
func NewFavoriteRepository(db *sql.DB, schema string, logger *zap.Logger) *FavoriteRepository {
	return &FavoriteRepository{
		db:     db,
		schema: schema,
		logger: logger,
	}
}

// GetByID получает избранное по ID
func (r *FavoriteRepository) GetByID(ctx context.Context, id string) (*model.Favorite, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_id, entity_type, entity_id, entity_name,
			notes, created_at
		FROM %s.favorites
		WHERE id = $1
	`, r.schema)

	var fav model.Favorite
	var entityType string
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&fav.ID,
		&fav.UserID,
		&entityType,
		&fav.EntityID,
		&fav.EntityName,
		&notes,
		&fav.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite: %w", err)
	}

	// Конвертируем entityType обратно в uppercase для соответствия GraphQL enum
	fav.EntityType = model.EntityType(strings.ToUpper(entityType))

	if notes.Valid {
		fav.Notes = &notes.String
	}

	return &fav, nil
}

// GetByUserID получает все избранное пользователя
func (r *FavoriteRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Favorite, error) {
	query := fmt.Sprintf(`
		SELECT
			id, user_id, entity_type, entity_id, entity_name,
			notes, created_at
		FROM %s.favorites
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*model.Favorite

	for rows.Next() {
		var fav model.Favorite
		var entityType string
		var notes sql.NullString

		err := rows.Scan(
			&fav.ID,
			&fav.UserID,
			&entityType,
			&fav.EntityID,
			&fav.EntityName,
			&notes,
			&fav.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}

		// Конвертируем entityType обратно в uppercase для соответствия GraphQL enum
		fav.EntityType = model.EntityType(strings.ToUpper(entityType))

		if notes.Valid {
			fav.Notes = &notes.String
		}

		favorites = append(favorites, &fav)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, nil
}

// HasFavorite проверяет наличие в избранном
func (r *FavoriteRepository) HasFavorite(ctx context.Context, userID, entityType, entityID string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s.favorites
			WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3
		)
	`, r.schema)

	// Конвертируем в lowercase для соответствия значениям в БД
	entityType = strings.ToLower(entityType)

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, entityType, entityID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite existence: %w", err)
	}

	return exists, nil
}

// Create создает новое избранное
func (r *FavoriteRepository) Create(ctx context.Context, favorite *model.Favorite) error {
	// Генерируем ID если не задан
	if favorite.ID == "" {
		favorite.ID = uuid.New().String()
	}

	favorite.CreatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s.favorites (
			id, user_id, entity_type, entity_id, entity_name,
			notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, r.schema)

	// Конвертируем EntityType в lowercase для соответствия check constraint
	entityType := strings.ToLower(string(favorite.EntityType))

	_, err := r.db.ExecContext(ctx, query,
		favorite.ID,
		favorite.UserID,
		entityType,
		favorite.EntityID,
		favorite.EntityName,
		favorite.Notes,
		favorite.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create favorite: %w", err)
	}

	r.logger.Info("favorite created",
		zap.String("id", favorite.ID),
		zap.String("user_id", favorite.UserID),
		zap.String("entity_type", string(favorite.EntityType)),
		zap.String("entity_id", favorite.EntityID),
	)

	return nil
}

// Update обновляет избранное (только notes)
func (r *FavoriteRepository) Update(ctx context.Context, favorite *model.Favorite) error {
	query := fmt.Sprintf(`
		UPDATE %s.favorites
		SET notes = $1
		WHERE id = $2
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query,
		favorite.Notes,
		favorite.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("favorite not found: %s", favorite.ID)
	}

	r.logger.Info("favorite updated",
		zap.String("id", favorite.ID),
	)

	return nil
}

// Delete удаляет избранное
func (r *FavoriteRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s.favorites
		WHERE id = $1
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("favorite not found: %s", id)
	}

	r.logger.Info("favorite deleted",
		zap.String("id", id),
	)

	return nil
}
