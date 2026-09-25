package model

import "time"

// Cursor menambatkan posisi pada baris TERAKHIR yang sudah dikirim,
// bukan menghitung "lewati N baris" seperti OFFSET.
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

// CursorQuery adalah parameter permintaan GET /students dengan pagination cursor.
type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

// CursorMeta menggantikan Meta pada endpoint yang memakai cursor.
//
// Tidak ada Total / TotalPages di sini: keduanya butuh COUNT(*) atas
// seluruh tabel — persis biaya yang ingin dihindari oleh cursor pagination.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
