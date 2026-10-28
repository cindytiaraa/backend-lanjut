# Tugas 10: Portofolio Mahasiswa dengan MongoDB

Nama / NIM : ...................
Repository : (isi tautan)

> Bagian bertanda **[ISI]** harus diisi dari hasil menjalankan aplikasi di komputer
> Anda sendiri. Keluaran di bawah tidak boleh disalin dari tempat lain.

## 1. Rancangan dokumen

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

## 2. Potongan kode penting

- `app/repository/portfolio_repository.go`: `AddProject` (`$push` + filter `projects.19`), `RemoveProject` (`$pull`), `TopTechnologies` (dua `$unwind`), `Replace` (upsert + ulang sekali bila duplicate key).
- `app/service/portfolio_service.go`: pemeriksaan student di PostgreSQL dan hak akses.
- `app/model/portfolio.go`: model tanpa tipe driver MongoDB.

## 3. Bukti (C.3)

1. Keluaran curl seluruh endpoint + 3 kasus gagal berbeda: **[ISI]** (`./scripts/uji_portofolio.sh | tee bukti_curl.txt`)
2. Batas 20 proyek serentak (18 terisi, 10 permintaan, tepat 2 berhasil): **[ISI]**
3. NoSQL injection ditolak: **[ISI]** (bagian "NoSQL INJECTION" pada skrip yang sama)
4. `explain("executionStats")` : **[ISI]** (`mongosh scripts/explain_portofolio.js`)
5. Tangkapan layar Compass satu dokumen portofolio lengkap: **[ISI]**

## 4. Analisis singkat (C.4): draf, sesuaikan dengan kata-kata Anda

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
`/prestasi`) tetap berfungsi; endpoint portofolio mengembalikan 503 untuk
galat koneksi, pemilihan server, atau timeout MongoDB, dan `/health` menjadi
503. Batas waktu request tetap 5 detik, sehingga client dapat mencoba kembali
setelah layanan pulih.
