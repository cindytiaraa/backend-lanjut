# Backend Lanjut — Layanan Akademik

## Gambaran proyek

Proyek ini adalah layanan backend untuk mendukung pengelolaan data akademik mahasiswa. API dibangun dengan Go dan Fiber. PostgreSQL menyimpan data inti (akun, mahasiswa, kelas dan KRS, pembimbing); MongoDB menyimpan portofolio mahasiswa berbentuk dokumen.

## Cakupan layanan

- **Akun dan akses:** pendaftaran, login, token akses, serta pembatasan akses berdasarkan peran dan izin.
- **Data mahasiswa:** pengelolaan informasi mahasiswa dengan validasi dan perlindungan akses.
- **Prestasi mahasiswa:** pencatatan dan pengambilan data prestasi.
- **Kelas dan KRS:** melihat kelas, mendaftar atau membatalkan kelas, serta berpindah kelas dengan perlindungan terhadap kuota dan batas SKS.
- **Pembimbing tugas akhir:** melihat pembimbing, menetapkan atau mengganti pembimbing, serta mengelola kuota bimbingan bagi pengguna berizin.
- **Portofolio mahasiswa (MongoDB):** satu dokumen per mahasiswa berisi proyek dan sertifikat, batas 20 proyek yang dijaga atomik, dan statistik 10 teknologi terpopuler.

## Keandalan dan perlindungan data

Perubahan KRS dan penetapan pembimbing menggunakan transaksi agar satu rangkaian perubahan tersimpan utuh atau dibatalkan seluruhnya ketika terjadi kegagalan. Pengambilan kuota dilakukan secara aman saat ada permintaan serentak. Penguncian data dan pemeriksaan versi membantu mencegah perubahan yang saling menimpa. Basis data juga menerapkan batasan untuk menjaga jumlah peserta kelas dan bimbingan tetap sesuai kuota.

## Status verifikasi

Pemeriksaan build, analisis statis, dan unit test Go telah lulus pada source saat ini. Hasil uji beban PostgreSQL yang tercantum dalam dokumen pengerjaan sebelumnya adalah hasil historis; pengujian beban belum diulang setelah perubahan terakhir. Karena itu, validasi akhir atas skenario transaksi dan konkurensi masih memerlukan database uji terpisah.


## Menjalankan MongoDB (Modul 10)

1. Pasang MongoDB (lihat Modul 10, Langkah 1) atau pakai MongoDB Atlas.
2. Isi `MONGO_URI` dan `MONGO_DATABASE` di `.env` (contoh di `.env.example`).
3. Jalankan aplikasi. Index MongoDB untuk portofolio dibuat otomatis saat aplikasi menyala.
4. Driver MongoDB v2 sudah tercantum di `go.mod` dan `go.sum`; tidak perlu menambahkannya lagi.

Endpoint baru: `/api/v1/students/:id/portfolio` (termasuk pengelolaan `/projects`) dan `/api/v1/portfolios/stats/technologies`. Skrip bukti portofolio ada di `scripts/` dan mewajibkan `ALLOW_PORTFOLIO_MUTATION=1` karena mengubah data student yang dipilih.
