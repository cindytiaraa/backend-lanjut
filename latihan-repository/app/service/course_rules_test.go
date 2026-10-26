package service

import (
	"testing"

	"latihan-repository/app/model"
)

// Pengujian ini tidak menyalakan server, tidak menyentuh database, dan
// tidak membuat fiber.Ctx.

func TestWithinCreditLimit(t *testing.T) {
	tests := []struct {
		name          string
		taken, adding int
		want          bool
	}{
		{"kosong", 0, 3, true},
		{"tepat di batas", 21, 3, true},
		{"melebihi satu SKS", 22, 3, false},
		{"sudah penuh", 24, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinCreditLimit(tt.taken, tt.adding); got != tt.want {
				t.Errorf("WithinCreditLimit(%d,%d)=%v, want %v", tt.taken, tt.adding, got, tt.want)
			}
		})
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	if got := CreditsAfterSwap(24, 3, 3); got != 24 {
		t.Errorf("swap setara SKS: got %d, want 24", got)
	}
	if got := CreditsAfterSwap(21, 2, 6); got != 25 {
		t.Errorf("swap menambah SKS: got %d, want 25", got)
	}
}

func TestTotalCredits(t *testing.T) {
	if got := TotalCredits(nil); got != 0 {
		t.Errorf("KRS kosong: got %d, want 0", got)
	}
	krs := []model.Course{{Credits: 3}, {Credits: 2}, {Credits: 4}}
	if got := TotalCredits(krs); got != 9 {
		t.Errorf("got %d, want 9", got)
	}
}

func TestValidateSwap(t *testing.T) {
	if errs := ValidateSwap(model.SwapRequest{FromCourseID: 1, ToCourseID: 2}); errs != nil {
		t.Errorf("kelas berbeda seharusnya lolos, dapat %v", errs)
	}
	errs := ValidateSwap(model.SwapRequest{FromCourseID: 5, ToCourseID: 5})
	if errs["to_course_id"] == "" {
		t.Error("kelas asal sama dengan tujuan seharusnya ditolak pada to_course_id")
	}
}
