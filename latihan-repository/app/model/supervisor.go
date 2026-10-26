package model

import "time"

type Supervisor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Quota           int       `json:"quota"`
	SupervisedCount int       `json:"supervised_count"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
}

// AssignRequest: POST /supervisors/:id/assign
type AssignRequest struct {
	StudentID int `json:"student_id" validate:"required,min=1"`
}

// SwitchSupervisionRequest: POST /supervisions/switch.
// Pembimbing lama tidak dikirim client; ia dibaca dari database di dalam
// transaction (setelah baris student dikunci), supaya tidak bisa dipalsukan
// dan tidak basi.
type SwitchSupervisionRequest struct {
	StudentID      int `json:"student_id" validate:"required,min=1"`
	ToSupervisorID int `json:"to_supervisor_id" validate:"required,min=1"`
}

// UpdateSupervisorQuotaRequest: PATCH /supervisors/:id (optimistic locking).
type UpdateSupervisorQuotaRequest struct {
	Quota   int `json:"quota" validate:"required,min=1,max=100"`
	Version int `json:"version" validate:"required,min=1"`
}
