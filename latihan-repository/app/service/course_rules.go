package service

import "latihan-repository/app/model"

// MaxCreditsPerSemester adalah batas SKS yang boleh diambil dalam satu KRS.
const MaxCreditsPerSemester = 24

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx, tidak
// menyentuh database, dan tidak tahu apa pun tentang HTTP. Seluruh isinya
// dapat diuji hanya dengan memanggilnya.

// WithinCreditLimit menjawab apakah menambah "adding" SKS pada "taken" SKS
// masih dalam batas.
//
// Aturan yang benar ini TIDAK cukup untuk menjamin batas SKS. Ia hanya
// menjawab benar bila angka "taken" yang diberikan kepadanya masih berlaku
// saat keputusan disimpan. Yang menjamin hal itu adalah kunci LockStudent,
// bukan fungsi ini.
func WithinCreditLimit(taken, adding int) bool {
	return taken+adding <= MaxCreditsPerSemester
}

// CreditsAfterSwap menghitung SKS setelah pindah kelas.
func CreditsAfterSwap(taken, leaving, joining int) int {
	return taken - leaving + joining
}

// TotalCredits menjumlahkan SKS seluruh kelas pada sebuah KRS.
func TotalCredits(courses []model.Course) int {
	total := 0
	for _, course := range courses {
		total += course.Credits
	}
	return total
}

// ValidateSwap memeriksa aturan antar-field yang tidak dapat ditulis sebagai
// tag: kelas asal dan kelas tujuan harus berbeda. Mengembalikan nil bila
// lolos, mengikuti bentuk helper.ValidateStruct.
func ValidateSwap(req model.SwapRequest) map[string]string {
	if req.FromCourseID == req.ToCourseID {
		return map[string]string{
			"to_course_id": "tidak boleh sama dengan from_course_id",
		}
	}
	return nil
}
