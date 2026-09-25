package service

import (
	"strings"

	"latihan-repository/app/model"
)

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada.
// Field bernilai nil dibiarkan apa adanya.
//
// Sejak validasi pindah ke tag (helper.ValidateStruct dipanggil di service
// SEBELUM fungsi ini), pemeriksaan bentuk sudah selesai. Tugas ApplyPatch
// tinggal satu: menggabungkan. Karena itu ia tidak lagi mengembalikan error.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}

	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}

	if req.Grade != nil {
		current.Grade = *req.Grade
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current
}

// IsEmptyPatch memeriksa body PATCH yang tidak berisi field apa pun.
//
// Aturan ini tidak dapat ditulis sebagai tag: tag memeriksa satu field pada
// satu waktu, sedangkan aturan ini berbicara tentang HUBUNGAN antar field —
// setidaknya satu di antaranya harus ada.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}
