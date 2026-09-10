package model

type Prestasi struct {
	IDPrestasi   int    `json:"id_prestasi"`
	IDStudent    int    `json:"id_student"`
	NamaPrestasi string `json:"nama_prestasi"`
	Juara        string `json:"juara"`
}

type CreatePrestasiRequest struct {
	IDStudent    int    `json:"id_student"`
	NamaPrestasi string `json:"nama_prestasi"`
	Juara        string `json:"juara"`
}
