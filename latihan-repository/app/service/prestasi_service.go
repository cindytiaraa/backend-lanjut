package service

import (
	"context"
	"errors"
	"strings"

	"latihan-repository/app/model"
	"latihan-repository/app/repository"
)

type PrestasiService struct {
	repo repository.PrestasiRepository
}

func NewPrestasiService(repo repository.PrestasiRepository) *PrestasiService {
	return &PrestasiService{repo: repo}
}

func (s *PrestasiService) List(ctx context.Context) ([]model.Prestasi, error) {
	return s.repo.FindAll(ctx)
}

func (s *PrestasiService) Get(ctx context.Context, id int) (model.Prestasi, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PrestasiService) Create(
	ctx context.Context,
	req model.CreatePrestasiRequest,
) (model.Prestasi, map[string]string, error) {

	errs := validatePrestasi(req)

	if len(errs) > 0 {
		return model.Prestasi{}, errs, nil
	}

	req.NamaPrestasi = strings.TrimSpace(req.NamaPrestasi)
	req.Juara = strings.TrimSpace(req.Juara)

	result, err := s.repo.Create(ctx, req)
	if err != nil {
		return model.Prestasi{}, nil, err
	}

	return result, nil, nil
}

func (s *PrestasiService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func validatePrestasi(req model.CreatePrestasiRequest) map[string]string {
	errs := map[string]string{}

	if req.IDStudent < 1 {
		errs["id_student"] = "wajib diisi"
	}

	if strings.TrimSpace(req.NamaPrestasi) == "" {
		errs["nama_prestasi"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Juara) == "" {
		errs["juara"] = "wajib diisi"
	}

	return errs
}

func IsPrestasiNotFound(err error) bool {
	return errors.Is(err, repository.ErrPrestasiNotFound)
}
