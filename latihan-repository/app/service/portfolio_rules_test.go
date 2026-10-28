package service

import (
	"reflect"
	"testing"

	"latihan-repository/app/model"
)

func TestNormalizeProject(t *testing.T) {
	got := normalizeProject(model.ProjectInput{
		Title:        "  API Kampus  ",
		Technologies: []string{"Go", " go ", "PostgreSQL", ""},
		URL:          " https://github.com/x/y ",
		Year:         2025,
	})

	want := model.ProjectInput{
		Title:        "API Kampus",
		Technologies: []string{"go", "postgresql"},
		URL:          "https://github.com/x/y",
		Year:         2025,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("normalizeProject = %+v, diharapkan %+v", got, want)
	}
}

func TestNormalizePortfolioRequest(t *testing.T) {
	got := normalizePortfolioRequest(model.ReplacePortfolioRequest{
		Headline:     " Backend Dev ",
		Bio:          "  halo  ",
		Projects:     []model.ProjectInput{{Title: " A ", Technologies: []string{"Go", "GO"}}},
		Certificates: []model.CertificateInput{{Name: " Sertif ", Issuer: " Dicoding ", IssuedAt: " 2025-03-01 "}},
	})

	if got.Headline != "Backend Dev" || got.Bio != "halo" {
		t.Errorf("headline/bio belum dirapikan: %+v", got)
	}
	if !reflect.DeepEqual(got.Projects[0].Technologies, []string{"go"}) {
		t.Errorf("teknologi = %q", got.Projects[0].Technologies)
	}
	if got.Certificates[0].IssuedAt != "2025-03-01" || got.Certificates[0].Issuer != "Dicoding" {
		t.Errorf("sertifikat belum dirapikan: %+v", got.Certificates[0])
	}
}
