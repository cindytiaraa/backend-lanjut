package service

import (
	"latihan-repository/app/model"
	"strings"
)

// ValidateCreate memeriksa isi permintaan pembuatan student.
// Mengembalikan peta berisi field yang bermasalah;
// kosong berarti lolos.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0-100"
	}

	return errs
}

// ValidateReplace memeriksa isi permintaan PUT.
// Seluruh field wajib ada karena PUT mengganti isi secara keseluruhan.
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus antara 0-100 pada PUT"
	}

	return errs
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field yang bernilai nil dibiarkan apa adanya.
func ApplyPatch(
	current model.Student,
	req model.PatchStudentRequest,
) (model.Student, map[string]string) {

	errs := map[string]string{}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
			return current, errs
		}

		current.NIM = *req.NIM
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
			return current, errs
		}

		current.Name = *req.Name
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus antara 0-100"
			return current, errs
		}
		current.Grade = *req.Grade
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
