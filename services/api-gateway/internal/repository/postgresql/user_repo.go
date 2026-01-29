package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// User представляет пользователя системы
type User struct {
	ID                       string
	Email                    string
	PasswordHash             string
	FirstName                string
	LastName                 string
	IsActive                 bool
	EmailVerified            bool
	EmailVerificationToken   *string
	EmailVerificationExpires *time.Time
	PasswordResetToken       *string
	PasswordResetExpires     *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
	LastLoginAt              *time.Time
}

// UserRepository реализация для работы с пользователями
type UserRepository struct {
	db     *sql.DB
	schema string
	logger *zap.Logger
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB, schema string, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		schema: schema,
		logger: logger,
	}
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := fmt.Sprintf(`
		SELECT
			id, email, password_hash, first_name, last_name,
			is_active, email_verified,
			email_verification_token, email_verification_expires_at,
			password_reset_token, password_reset_expires_at,
			created_at, updated_at, last_login_at
		FROM %s.users
		WHERE id = $1
	`, r.schema)

	var user User
	var emailVerificationToken, passwordResetToken sql.NullString
	var emailVerificationExpires, passwordResetExpires, lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.EmailVerified,
		&emailVerificationToken,
		&emailVerificationExpires,
		&passwordResetToken,
		&passwordResetExpires,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to get user by id", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Обработка nullable полей
	if emailVerificationToken.Valid {
		user.EmailVerificationToken = &emailVerificationToken.String
	}
	if emailVerificationExpires.Valid {
		user.EmailVerificationExpires = &emailVerificationExpires.Time
	}
	if passwordResetToken.Valid {
		user.PasswordResetToken = &passwordResetToken.String
	}
	if passwordResetExpires.Valid {
		user.PasswordResetExpires = &passwordResetExpires.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf(`
		SELECT
			id, email, password_hash, first_name, last_name,
			is_active, email_verified,
			email_verification_token, email_verification_expires_at,
			password_reset_token, password_reset_expires_at,
			created_at, updated_at, last_login_at
		FROM %s.users
		WHERE email = $1
	`, r.schema)

	var user User
	var emailVerificationToken, passwordResetToken sql.NullString
	var emailVerificationExpires, passwordResetExpires, lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.EmailVerified,
		&emailVerificationToken,
		&emailVerificationExpires,
		&passwordResetToken,
		&passwordResetExpires,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Обработка nullable полей
	if emailVerificationToken.Valid {
		user.EmailVerificationToken = &emailVerificationToken.String
	}
	if emailVerificationExpires.Valid {
		user.EmailVerificationExpires = &emailVerificationExpires.Time
	}
	if passwordResetToken.Valid {
		user.PasswordResetToken = &passwordResetToken.String
	}
	if passwordResetExpires.Valid {
		user.PasswordResetExpires = &passwordResetExpires.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	// Генерируем UUID если не указан
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := fmt.Sprintf(`
		INSERT INTO %s.users (
			id, email, password_hash, first_name, last_name,
			is_active, email_verified,
			email_verification_token, email_verification_expires_at,
			password_reset_token, password_reset_expires_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, r.schema)

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
		user.EmailVerificationToken,
		user.EmailVerificationExpires,
		user.PasswordResetToken,
		user.PasswordResetExpires,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create user", zap.Error(err), zap.String("email", user.Email))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("user created", zap.String("id", user.ID), zap.String("email", user.Email))
	return nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET
			first_name = $1,
			last_name = $2,
			is_active = $3,
			email_verified = $4,
			email_verification_token = $5,
			email_verification_expires_at = $6,
			password_reset_token = $7,
			password_reset_expires_at = $8,
			updated_at = $9
		WHERE id = $10
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
		user.EmailVerificationToken,
		user.EmailVerificationExpires,
		user.PasswordResetToken,
		user.PasswordResetExpires,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		r.logger.Error("failed to update user", zap.Error(err), zap.String("id", user.ID))
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("user updated", zap.String("id", user.ID))
	return nil
}

// UpdatePassword обновляет пароль пользователя
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := fmt.Sprintf(`
		UPDATE %s.users
		SET
			password_hash = $1,
			password_reset_token = NULL,
			password_reset_expires_at = NULL,
			updated_at = $2
		WHERE id = $3
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		r.logger.Error("failed to update password", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("password updated", zap.String("user_id", userID))
	return nil
}

// UpdateLastLogin обновляет время последнего входа
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	query := fmt.Sprintf(`
		UPDATE %s.users
		SET last_login_at = $1
		WHERE id = $2
	`, r.schema)

	_, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		r.logger.Error("failed to update last login", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// Delete удаляет пользователя (soft delete через is_active = false)
func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := fmt.Sprintf(`
		UPDATE %s.users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, r.schema)

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		r.logger.Error("failed to delete user", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("user deleted (soft)", zap.String("user_id", userID))
	return nil
}
