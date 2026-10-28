package service

import (
	"reflect"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil menjadi slice kosong", nil, []string{}},
		{"huruf kecil dan trim", []string{" Tugas-Banyak "}, []string{"tugas-banyak"}},
		{"duplikat dibuang, urutan pertama tetap", []string{"Go", "sql", "go", " SQL"}, []string{"go", "sql"}},
		{"kosong dibuang", []string{"", "   ", "go"}, []string{"go"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeTags(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NormalizeTags(%q) = %q, diharapkan %q", tt.in, got, tt.want)
			}
		})
	}
}
