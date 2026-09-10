package service

import (
	"latihan-repository/app/model"
	"testing"
)

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:   "",
		Name:  "",
		Grade: 120,
	}

	errs := ValidateCreate(req)

	if len(errs) != 3 {
		t.Fatalf(
			"seharusnya ada 3 error, dapat %d: %v",
			len(errs),
			errs,
		)
	}
}

func TestValidateReplace(t *testing.T) {
	req := model.ReplaceStudentRequest{
		NIM:   "",
		Name:  "",
		Grade: -10,
	}

	errs := ValidateReplace(req)

	if len(errs) != 3 {
		t.Fatalf(
			"seharusnya ada 3 error, dapat %d: %v",
			len(errs),
			errs,
		)
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "001",
		Name:     "Cindy",
		Grade:    80,
		IsActive: true,
	}

	newGrade := 90.0

	result, errs := ApplyPatch(
		initial,
		model.PatchStudentRequest{
			Grade: &newGrade,
		},
	)

	if len(errs) != 0 {
		t.Fatalf(
			"tidak seharusnya ada error: %v",
			errs,
		)
	}

	if result.Grade != 90 {
		t.Errorf(
			"grade seharusnya berubah menjadi 90, dapat %v",
			result.Grade,
		)
	}

	if result.Name != "Cindy" {
		t.Error(
			"field yang tidak dikirim seharusnya tidak berubah",
		)
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total int
		limit int
		want  int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}

	for _, tc := range cases {
		got := CountTotalPages(
			tc.total,
			tc.limit,
		)

		if got != tc.want {
			t.Errorf(
				"total=%d limit=%d: harap %d, dapat %d",
				tc.total,
				tc.limit,
				tc.want,
				got,
			)
		}
	}
}
