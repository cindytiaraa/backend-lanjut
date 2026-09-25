package service

import (
	"testing"

	"latihan-repository/app/model"
)

// Validasi manual (ValidateCreate/ValidateReplace) sudah dihapus sejak
// Modul 7 — digantikan tag pada struct + helper.ValidateStruct. Pengujian
// aturan tag dilakukan lewat helper.ValidateStruct, bukan di sini lagi.

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "20230001",
		Name:     "Cindy",
		Grade:    80,
		IsActive: true,
	}

	newGrade := 90.0

	result := ApplyPatch(
		initial,
		model.PatchStudentRequest{
			Grade: &newGrade,
		},
	)

	if result.Grade != 90 {
		t.Errorf("grade seharusnya berubah menjadi 90, dapat %v", result.Grade)
	}

	if result.Name != "Cindy" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("body kosong seharusnya dianggap empty patch")
	}

	name := "Budi"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &name}) {
		t.Error("body berisi satu field seharusnya TIDAK dianggap empty patch")
	}
}
