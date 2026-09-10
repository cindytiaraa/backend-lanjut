package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-repository/app/model"
)

var ErrPrestasiNotFound = errors.New("prestasi tidak ditemukan")

type PrestasiRepository interface {
	FindAll(ctx context.Context) ([]model.Prestasi, error)
	FindByID(ctx context.Context, id int) (model.Prestasi, error)
	Create(ctx context.Context, req model.CreatePrestasiRequest) (model.Prestasi, error)
	Delete(ctx context.Context, id int) error
}

type prestasiRepository struct {
	pool *pgxpool.Pool
}

func NewPrestasiRepository(pool *pgxpool.Pool) PrestasiRepository {
	return &prestasiRepository{pool: pool}
}

func (r *prestasiRepository) FindAll(ctx context.Context) ([]model.Prestasi, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id_prestasi, id_student, nama_prestasi, juara
		FROM prestasi
		ORDER BY id_prestasi
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Prestasi

	for rows.Next() {
		var p model.Prestasi

		if err := rows.Scan(
			&p.IDPrestasi,
			&p.IDStudent,
			&p.NamaPrestasi,
			&p.Juara,
		); err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *prestasiRepository) FindByID(ctx context.Context, id int) (model.Prestasi, error) {
	var p model.Prestasi

	err := r.pool.QueryRow(ctx, `
		SELECT id_prestasi, id_student, nama_prestasi, juara
		FROM prestasi
		WHERE id_prestasi = $1
	`, id).Scan(
		&p.IDPrestasi,
		&p.IDStudent,
		&p.NamaPrestasi,
		&p.Juara,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Prestasi{}, ErrPrestasiNotFound
	}

	if err != nil {
		return model.Prestasi{}, err
	}

	return p, nil
}

func (r *prestasiRepository) Create(ctx context.Context, req model.CreatePrestasiRequest) (model.Prestasi, error) {
	var p model.Prestasi

	err := r.pool.QueryRow(ctx, `
		INSERT INTO prestasi (id_student, nama_prestasi, juara)
		VALUES ($1, $2, $3)
		RETURNING id_prestasi, id_student, nama_prestasi, juara
	`,
		req.IDStudent,
		req.NamaPrestasi,
		req.Juara,
	).Scan(
		&p.IDPrestasi,
		&p.IDStudent,
		&p.NamaPrestasi,
		&p.Juara,
	)

	if err != nil {
		return model.Prestasi{}, err
	}

	return p, nil
}

func (r *prestasiRepository) Delete(ctx context.Context, id int) error {
	commandTag, err := r.pool.Exec(ctx, `
		DELETE FROM prestasi
		WHERE id_prestasi = $1
	`, id)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrPrestasiNotFound
	}

	return nil
}
