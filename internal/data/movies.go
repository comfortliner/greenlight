package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/comfortliner/greenlight/internal/validator"
)

// **********************
// * Model Definition
// **********************

// Define a Movie struct to represent an individual movie.
type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie.
	CreatedAt time.Time `json:"-"`                 // Timestamp for when the movie is added to our database.
	Title     string    `json:"title"`             // Movie title.
	Year      int32     `json:"year,omitempty"`    // Movie release year.
	Runtime   int32     `json:"runtime,omitempty"` // Movie runtime (in minutes).
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each time the movie information is updated.
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// **********************
// * Data Validation
// **********************

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more then 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be an positive integer")
}

// **********************
// * Data Manipulation
// **********************

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime)
		OUTPUT INSERTED.id, INSERTED.created_at, INSERTED.version
		VALUES (@p1, @p2, @p3);
	`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Version,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime,  version
		FROM movies
		WHERE id = @p1;
	`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) GetAll(title string, filters Filters) ([]*Movie, error) {
	query := fmt.Sprintf(`
		SELECT id, created_at, title, year, runtime, version
		FROM movies
		WHERE (LOWER(title) = LOWER(@p1) OR @p1 = '')
		ORDER BY %s %s;
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	movies := []*Movie{}

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies 
		SET title = @p1, year = @p2, runtime = @p3, version = version + 1
		OUTPUT INSERTED.version
		WHERE id = @p4;
	`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = @p1
	`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
