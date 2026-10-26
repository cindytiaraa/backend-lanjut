package service

// File ini berisi business rules MURNI untuk pembimbing tugas akhir: tidak
// menyentuh fiber.Ctx dan tidak menyentuh database.

// IsSameSupervisor menandai pergantian pembimbing yang tidak bermakna
// (pembimbing tujuan sama dengan pembimbing saat ini).
func IsSameSupervisor(currentID, targetID int) bool {
	return currentID == targetID
}
