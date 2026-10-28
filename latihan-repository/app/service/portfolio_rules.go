package service

import (
	"strings"

	"latihan-repository/app/model"
)

// normalizeProject merapikan satu isian proyek sebelum divalidasi: spasi di
// tepi dibuang, teknologi dijadikan huruf kecil tanpa duplikat. Tanpa ini
// "Go", "go", dan " GO " terhitung tiga teknologi berbeda di statistik.
func normalizeProject(p model.ProjectInput) model.ProjectInput {
	p.Title = strings.TrimSpace(p.Title)
	p.URL = strings.TrimSpace(p.URL)
	p.Technologies = NormalizeTags(p.Technologies)
	return p
}

func normalizeCertificate(c model.CertificateInput) model.CertificateInput {
	c.Name = strings.TrimSpace(c.Name)
	c.Issuer = strings.TrimSpace(c.Issuer)
	c.IssuedAt = strings.TrimSpace(c.IssuedAt)
	return c
}

func normalizePortfolioRequest(r model.ReplacePortfolioRequest) model.ReplacePortfolioRequest {
	r.Headline = strings.TrimSpace(r.Headline)
	r.Bio = strings.TrimSpace(r.Bio)
	for i := range r.Projects {
		r.Projects[i] = normalizeProject(r.Projects[i])
	}
	for i := range r.Certificates {
		r.Certificates[i] = normalizeCertificate(r.Certificates[i])
	}
	return r
}
