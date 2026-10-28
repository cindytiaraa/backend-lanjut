package model

import "time"

// Portfolio adalah bentuk portofolio yang dikenal seluruh aplikasi.
//
// Seperti model.Review, tidak ada tipe driver MongoDB di sini: ID berupa
// string. Terjemahan ke BSON adalah urusan repository.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Project disimpan DI DALAM dokumen portofolio (embedding).
type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

// Certificate juga di-embed. IssuedAt berupa tanggal "YYYY-MM-DD".
type Certificate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Issuer   string `json:"issuer"`
	IssuedAt string `json:"issued_at"`
}

// ProjectInput adalah isian satu proyek, dipakai oleh PUT portofolio maupun
// POST proyek. ID dibuat oleh repository, bukan dikirim client.
type ProjectInput struct {
	Title        string   `json:"title" validate:"required,min=3,max=100"`
	Technologies []string `json:"technologies" validate:"max=10,dive,min=1,max=30"`
	URL          string   `json:"url" validate:"omitempty,max=300,httpurl"`
	Year         int      `json:"year" validate:"required,min=1990,max=2100"`
}

type CertificateInput struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Issuer   string `json:"issuer" validate:"required,min=2,max=100"`
	IssuedAt string `json:"issued_at" validate:"required,datetime=2006-01-02"`
}

// ReplacePortfolioRequest dipakai PUT /students/:id/portfolio.
type ReplacePortfolioRequest struct {
	Headline     string             `json:"headline" validate:"required,min=3,max=100"`
	Bio          string             `json:"bio" validate:"max=1000"`
	Projects     []ProjectInput     `json:"projects" validate:"max=20,dive"`
	Certificates []CertificateInput `json:"certificates" validate:"max=30,dive"`
}

// TechnologyStat adalah satu baris peringkat teknologi.
type TechnologyStat struct {
	Technology string `json:"technology"`
	Projects   int    `json:"projects"`
	Students   int    `json:"students"`
}
