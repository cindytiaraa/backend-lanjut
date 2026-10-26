package service

import "testing"

func TestIsSameSupervisor(t *testing.T) {
	if !IsSameSupervisor(2, 2) {
		t.Error("pembimbing yang sama seharusnya terdeteksi")
	}
	if IsSameSupervisor(2, 3) {
		t.Error("pembimbing berbeda seharusnya tidak dianggap sama")
	}
}
