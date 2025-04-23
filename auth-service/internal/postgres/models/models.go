package models

import (
	"auth-service/api/calltypes"
	"auth-service/pkg/consts"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"auth-service/pkg/errormsg"
	"golang.org/x/crypto/bcrypt"
)

type PostgresRepository struct {
	Conn *sql.DB
}

func NewPostgresRepository(pool *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		Conn: pool,
	}
}

// UserExists checks does user really exist.
func (u *PostgresRepository) UserExists(id int) (bool, error) {
	var exists bool

	err := u.queryRow(context.Background(), "SELECT EXISTS(SELECT 1 from medods WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		log.Println("failed to check if user exists: ", err)

		return false, fmt.Errorf("failed to check user existence (id: %d): %w", id, err)
	}

	return exists, nil
}

// GetAll returns a slice of all users, sorted by last name.
func (u *PostgresRepository) GetAll() ([]*calltypes.User, error) {
	query := `select id, email, first_name, last_name, active, created_at, updated_at
              from medods`

	rows, err := u.Conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, errormsg.ErrFetchUser
	}
	defer rows.Close()

	var users []*calltypes.User

	for rows.Next() {
		var user calltypes.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error scanning user: %v", err)

			return nil, errormsg.ErrScanUser
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after row iteration: %v", err)

		return nil, errormsg.ErrFetchUser
	}

	return users, nil
}

// EmailCheck using to auth, gets password by provided email.
func (u *PostgresRepository) EmailCheck(email string) (*calltypes.User, error) {
	var emailExists bool

	err := u.queryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 from medods WHERE email = $1)", email).Scan(&emailExists)
	if err != nil {
		log.Println("failed to check email: ")

		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	if !emailExists {
		log.Println("User with that email does not exists")

		return nil, fmt.Errorf("user with that email does not exists: %w", err)
	}

	query := `select first_name, password from medods where email = $1`

	var user calltypes.User
	err = u.queryRow(context.Background(), query, email).Scan(
		&user.FirstName,
		&user.Password,
	)

	if err != nil {
		log.Println("failed to fetch user's password by email")

		return nil, fmt.Errorf("failed to fecth user's password by email: %w", err)
	}

	return &user, nil
}

// GetByEmail returns info of one user by email.
func (u *PostgresRepository) GetByEmail(email string) (*calltypes.User, error) {
	query := `select id, email, first_name, last_name, password, active, created_at, updated_at 
              from medods where email = $1`

	var user calltypes.User
	err := u.queryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		log.Println("failed to fetch user by email")

		return nil, fmt.Errorf("failed to fecth user's password by email: %w", err)
	}

	return &user, nil
}

// GetOne returns one user by id.
func (u *PostgresRepository) GetOne(id int) (*calltypes.User, error) {
	idExists, err := u.UserExists(id)
	if err != nil {
		return nil, err
	}

	if !idExists {
		return nil, errormsg.ErrUserNotFound
	}

	query := `select id, email, first_name, last_name, active, created_at, updated_at
              from medods where id = $1`

	var user calltypes.User
	err = u.queryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		log.Println("failed to fetch user by id: ", err)

		return nil, fmt.Errorf("failed to fetch user by id: %w", err)
	}

	return &user, nil
}

// Update updates one user in the database, using the information stored in the receiver u.
func (u *PostgresRepository) Update(user calltypes.User) error {
	idExists, err := u.UserExists(user.ID)
	if err != nil {
		return err
	}

	if !idExists {
		return errormsg.ErrUserNotFound
	}

	stmt := `update medods set
             email = $1,
             first_name = $2,
             last_name = $3,
             active = $4,
             updated_at = $5
             where id = $6`

	_, err = u.execQuery(context.Background(), stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Active,
		time.Now(),
		user.ID,
	)
	if err != nil {
		log.Println("failed to update user: ", err)

		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Insert adds new user to the database.
func (u *PostgresRepository) Insert(user calltypes.User) (int, error) {
	if len(user.Password) < consts.PassMinLength {
		return 0, errormsg.ErrPasswordLength
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), consts.BcryptCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	var newID int

	stmt := `insert into medods (email, first_name, last_name, password, active, created_at, updated_at)
         values ($1, $2, $3, $4, $5, $6, $7) returning id`

	err = u.queryRow(context.Background(), stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		hashedPassword,
		user.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		log.Println("failed to insert new user: ", err)

		return 0, fmt.Errorf("failed to insert new user: %w", err)
	}

	return newID, nil
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *PostgresRepository) PasswordMatches(plainText string, user calltypes.User) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, fmt.Errorf("failed to compare passwords: %w", err)
		}
	}

	return true, nil
}

// StoreRefreshToken stores provided refresh token.
func (u *PostgresRepository) StoreRefreshToken(id int, rawToken string) error {
	hashedToken, err := HashRefreshToken(rawToken)
	if err != nil {
		return err
	}

	stmt := `UPDATE medods SET refresh_token = $1, refresh_token_expires = $2 WHERE id = $3`

	_, err = u.execQuery(context.Background(), stmt,
		hashedToken,
		time.Now().Add(consts.RefreshTokenExpireTime),
		id,
	)

	return err
}

func (u *PostgresRepository) UpdateRefreshToken(id int, rawToken string) error {
	hashedToken, err := HashRefreshToken(rawToken)
	if err != nil {
		return err
	}

	stmt := `UPDATE medods SET refresh_token = $1, refresh_token_expires = $2 WHERE id = $3`
	_, err = u.execQuery(context.Background(), stmt,
		hashedToken,
		time.Now().Add(consts.RefreshTokenExpireTime),
		id,
	)

	return err
}

func HashRefreshToken(token string) (string, error) {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	return string(hashedToken), nil
}

func (u *PostgresRepository) ValidateRefreshToken(rawToken, clientIP string, id int) (bool, error) {
	parts := strings.Split(rawToken, "|")
	if len(parts) != consts.TokenParts {
		return false, errormsg.ErrInvalidRefreshToken
	}

	tokenIP := parts[0]
	if tokenIP != clientIP {
		var userEmail string
		err := u.queryRow(context.Background(),
			"SELECT email FROM medods WHERE id = $1", id).Scan(&userEmail)

		if err != nil {
			fmt.Printf("Failed to get user email for IP change warning: %v\n", err)

			userEmail = "mock_user@example.com"
		}

		// mock email warning
		warningMsg := fmt.Sprintf(
			"Security warning: Refresh attempt from new IP\n"+
				"Account ID: %d\n"+
				"Old IP: %s\n"+
				"New IP: %s\n"+
				"Time: %s",
			id, tokenIP, clientIP, time.Now().Format(time.RFC3339),
		)

		fmt.Printf("=== EMAIL WARNING ===\n"+
			"To: %s\n"+
			"Subject: Security Warning - New IP Detected\n"+
			"Body:\n%s\n"+
			"=====================\n",
			userEmail, warningMsg)

		return false, errormsg.ErrInvalidIP
	}

	var hashedToken string

	var expiresAt time.Time

	stmt := `SELECT refresh_token, refresh_token_expires FROM medods WHERE id = $1`

	err := u.queryRow(context.Background(), stmt, id).Scan(&hashedToken, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errormsg.ErrUserNotFound
		}
	}

	if time.Now().After(expiresAt) {
		return false, errormsg.ErrTokenExpired
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(rawToken))
	if err != nil {
		return false, errormsg.ErrCompareHash
	}

	return true, nil
}

func (u *PostgresRepository) execQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error) { //nolint: unparam
	ctx, cancel := context.WithTimeout(ctx, consts.DbTimeout)
	defer cancel()

	result, err := u.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query : %w", err)
	}

	return result, nil
}

func (u *PostgresRepository) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, consts.DbTimeout)
	defer cancel()

	return u.Conn.QueryRowContext(ctx, query, args...)
}
