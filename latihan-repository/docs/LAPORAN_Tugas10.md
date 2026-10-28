# Tugas 10: API dengan MongoDB

Nama / NIM : ...................
Repository : (isi tautan)

> Bagian bertanda **[ISI]** harus diisi dari hasil menjalankan aplikasi di komputer
> Anda sendiri. Keluaran di bawah tidak boleh disalin dari tempat lain.

## A. Ulasan kelas

Collection `course_reviews`, satu dokumen per ulasan. Balasan disimpan
tertanam pada dokumennya; `helpful_by` menyimpan user yang sudah memberi
penilaian agar hitungan tidak bisa dinaikkan berulang kali oleh pengguna yang
sama.

Index unik `course_user_unique` menjamin satu ulasan per user per kelas.
`course_newest` mendukung daftar berdasarkan kelas dan waktu terbaru, sedangkan
`course_tags` mendukung filter tag. Operasi balasan dan helpful menggunakan
pembaruan atomik MongoDB; balasan dibatasi 50 per ulasan.

Endpoint:

- `POST /api/v1/courses/:id/reviews` hanya mengizinkan peserta kelas.
- `GET /api/v1/courses/:id/reviews` mendukung pagination cursor dan filter tag.
- `GET /api/v1/courses/:id/reviews/summary` menghitung statistik dengan
  aggregation pipeline.
- `GET`, `PATCH`, dan `DELETE /api/v1/reviews/:id`; pemilik dapat mengubah dan
  menghapus ulasannya, sedangkan penghapusan ulasan siapa pun memerlukan
  permission `review:moderate`.
- `POST /api/v1/reviews/:id/replies` dan
  `POST /api/v1/reviews/:id/helpful`.

Untuk memberikan permission moderasi kepada role admin, jalankan
`migrations/010_review_permissions.sql`.

## B. Portofolio mahasiswa

Collection `student_portfolios`, satu dokumen per mahasiswa.

```json
{
  "_id": "ObjectId('...')",
  "student_id": 1,
  "headline": "Backend Developer",
  "bio": "Suka Go dan MongoDB",
  "projects": [
    {
      "_id": "ObjectId('...')",
      "title": "API Kampus",
      "technologies": ["go", "postgresql"],
      "url": "https://github.com/x/api-kampus",
      "year": 2025
    }
  ],
  "certificates": [
    {
      "_id": "ObjectId('...')",
      "name": "Belajar Backend",
      "issuer": "Dicoding",
      "issued_at": "ISODate('2025-03-01T00:00:00.000Z')"
    }
  ],
  "created_at": "ISODate('...')",
  "updated_at": "ISODate('...')"
}
```

Index: `student_unique` (unik pada `student_id`) menjaga "satu mahasiswa satu
portofolio" di level database. Tidak ada index lain karena seluruh endpoint
mencari dokumen lewat `student_id`; pipeline statistik membaca seluruh
collection sehingga index tidak membantunya.

### Potongan kode penting

- `app/repository/portfolio_repository.go`: `AddProject` (`$push` + filter `projects.19`), `RemoveProject` (`$pull`), `TopTechnologies` (dua `$unwind`), `Replace` (upsert + ulang sekali bila duplicate key).
- `app/service/portfolio_service.go`: pemeriksaan student di PostgreSQL dan hak akses.
- `app/model/portfolio.go`: model tanpa tipe driver MongoDB.
- `app/repository/review_repository.go`: index, pagination, pembaruan atomik,
  dan aggregation ulasan.
- `app/service/review_service.go`: validasi peserta kelas dan penerjemahan
  error dari MongoDB.
- `app/model/review.go`: model ulasan tanpa tipe driver MongoDB.

### Bukti portofolio (C.3)

1. Keluaran curl seluruh endpoint + 3 kasus gagal berbeda: **[ISI]** (`./scripts/uji_portofolio.sh | tee bukti_curl.txt`)
2. Batas 20 proyek serentak (18 terisi, 10 permintaan, tepat 2 berhasil): **[ISI]**
3. NoSQL injection ditolak: **[ISI]** (bagian "NoSQL INJECTION" pada skrip yang sama)
4. `explain("executionStats")` : **[ISI]** (`mongosh scripts/explain_portofolio.js`)
5. Tangkapan layar Compass satu dokumen portofolio lengkap: **[ISI]**

## C. Analisis singkat (C.4): draf, sesuaikan dengan kata-kata Anda

**10. Mengapa `projects` disarangkan?** Proyek hanya bermakna sebagai bagian dari
portofolio satu mahasiswa: selalu dibaca bersama portofolionya, tidak dipakai
dokumen lain, dan jumlahnya dibatasi 20 sehingga dokumen jauh di bawah batas
16 MB. Menyarangkannya membuat satu `GET` cukup satu dokumen, dan penambahan
proyek menjadi satu operasi atomik pada satu dokumen. Keputusan itu harus
dibalik bila proyek menjadi entitas yang berdiri sendiri, misalnya proyek
dikerjakan bersama beberapa mahasiswa, perlu dicari lintas mahasiswa dengan
filter rumit, atau batas 20 dicabut sehingga array bisa tumbuh tanpa batas.

**11. Mahasiswa dihapus dari PostgreSQL.** Tidak ada FOREIGN KEY lintas
database, sehingga dokumen portofolio tetap tinggal sebagai data yatim dan
masih ikut terhitung pada statistik teknologi. Cara 1: service penghapus
student juga menghapus portofolionya di MongoDB sesudah penghapusan di
PostgreSQL berhasil; kelemahannya tidak atomik, bila langkah kedua gagal data
yatim tetap ada. Cara 2: pekerjaan pembersih berkala yang mencari
`student_id` tanpa pasangan di PostgreSQL lalu menghapusnya; kelemahannya data
yatim hidup sampai pekerjaan berikutnya, dan menambah komponen yang harus
dipantau.

**12. Bila MongoDB mati.** Aplikasi menolak menyala bila MongoDB tidak bisa
di-ping saat start. Bila MongoDB mati ketika aplikasi sudah berjalan,
endpoint PostgreSQL (`/auth`, `/students`, `/courses`, `/supervisors`,
`/prestasi`) tetap berfungsi; endpoint MongoDB portofolio dan ulasan
mengembalikan 503 untuk
galat koneksi, pemilihan server, atau timeout MongoDB, dan `/health` menjadi
503. Batas waktu request tetap 5 detik, sehingga client dapat mencoba kembali
setelah layanan pulih.
