package data

import (
	"context"
	"time"

	"github.com/iosh/go-greenlight/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty,string"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {

	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *pgxpool.Pool
}

func (m MovieModel) Insert(movie *Movie) error {
	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres}
	return m.DB.QueryRow(context.Background(), "insert into movies(title, year, runtime, genres) values($1, $2,$3,$4) returning id, create_at, version", args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var movie Movie
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRow(ctx, "select id, create_at, title, year, runtime, genres, version from movies where id=$1", id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Genres,
		&movie.Version,
	); err != nil {
		return nil, err
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {

	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres, movie.ID, movie.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRow(
		ctx,
		"update movies set title=$1, year=$2, runtime=$3, genres=$4, version=version+1 where id=$5 and version=$6 returning version",
		args...,
	).Scan(&movie.Version); err != nil {
		return err
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, "delete from movies where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row, err := m.DB.Query(ctx, "select id, create_at, title, year, runtime, genres, version from movies where (LOWER(title) = LOWER($1) or $1='')AND (genres @> $2 OR $2 = '{}')  order by id", title, genres)

	if err != nil {
		return nil, err
	}

	movies, err := pgx.CollectRows(row, func(row pgx.CollectableRow) (*Movie, error) {
		var movie Movie
		e := row.Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, &movie.Genres, &movie.Version)
		return &movie, e
	})
	if err != nil {
		return nil, err
	}

	return movies, nil

}
