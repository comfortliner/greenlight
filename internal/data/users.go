package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/comfortliner/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

// **********************
// * Model Definition
// **********************

// Define a User struct to represent an individual user.
type User struct {
	ID        int64     `json:"id"`         // Unique integer ID for the user.
	CreatedAt time.Time `json:"created_at"` // Timestamp for when the user is added to our database.
	Name      string    `json:"name"`       // User name.
	Email     string    `json:"email"`      // User email.
	Password  password  `json:"-"`          // User password.
	Password2 password  `json:"-"`          // User password.
	Activated bool      `json:"activated"`  // User is activated [y/n].
	Version   int       `json:"-"`          // The version number starts at 1 and will be incremented each time the user information is updated.
}

// Define a UserModel struct type which wraps a sql.DB connection pool.
type UserModel struct {
	DB *sql.DB
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user.
type password struct {
	plaintext *string
	hash      []byte
}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// **********************
// * Data Validation
// **********************

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.plaintext != nil && user.Password2.plaintext != nil {
		ValidatePasswordMatchPlaintext(v, *user.Password.plaintext, *user.Password2.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidatePasswordMatchPlaintext(v *validator.Validator, password string, password2 string) {
	v.Check(password == password2, "password", "passwords do not match")
}

// **********************
// * Data Manipulation
// **********************

func (m UserModel) Insert(user *User) error {
	query := `
			INSERT INTO users (name, email, password_hash, activated)
			OUTPUT INSERTED.id, INSERTED.created_at, INSERTED.version
			VALUES (@p1, @p2, @p3, @p4);
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "mssql: Violation of UNIQUE KEY constraint"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
			SELECT id, created_at, name, email, password_hash, activated, version
			FROM users
			WHERE email = @p1;
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
			UPDATE users 
			SET name = @p1, email = @p2, password_hash = @p3, activated = @p4, version = version + 1
			OUTPUT INSERTED.version
			WHERE id = @p5 AND version = @p6;
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "mssql: Violation of UNIQUE KEY constraint"):
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		FROM users
		INNER JOIN usertokens
		ON users.id = usertokens.user_id
		WHERE usertokens.hash = @p1
		AND usertokens.scope = @p2
		AND usertokens.expiry > @p3
	`

	args := []any{
		tokenHash[:],
		tokenScope,
		time.Now(),
	}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
