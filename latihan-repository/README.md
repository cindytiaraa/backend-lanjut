# Backend Lanjut — Layanan Akademik

## Gambaran proyek

Proyek ini adalah layanan backend untuk mendukung pengelolaan data akademik mahasiswa. API dibangun dengan Go dan Fiber, menggunakan PostgreSQL sebagai penyimpanan data.

## Cakupan layanan

- **Akun dan akses:** pendaftaran, login, token akses, serta pembatasan akses berdasarkan peran dan izin.
- **Data mahasiswa:** pengelolaan informasi mahasiswa dengan validasi dan perlindungan akses.
- **Prestasi mahasiswa:** pencatatan dan pengambilan data prestasi.
- **Kelas dan KRS:** melihat kelas, mendaftar atau membatalkan kelas, serta berpindah kelas dengan perlindungan terhadap kuota dan batas SKS.
- **Pembimbing tugas akhir:** melihat pembimbing, menetapkan atau mengganti pembimbing, serta mengelola kuota bimbingan bagi pengguna berizin.

## Keandalan dan perlindungan data

Perubahan KRS dan penetapan pembimbing menggunakan transaksi agar satu rangkaian perubahan tersimpan utuh atau dibatalkan seluruhnya ketika terjadi kegagalan. Pengambilan kuota dilakukan secara aman saat ada permintaan serentak. Penguncian data dan pemeriksaan versi membantu mencegah perubahan yang saling menimpa. Basis data juga menerapkan batasan untuk menjaga jumlah peserta kelas dan bimbingan tetap sesuai kuota.

## Status verifikasi

Pemeriksaan build, analisis statis, dan unit test Go telah lulus pada source saat ini. Hasil uji beban PostgreSQL yang tercantum dalam dokumen pengerjaan sebelumnya adalah hasil historis; pengujian beban belum diulang setelah perubahan terakhir. Karena itu, validasi akhir atas skenario transaksi dan konkurensi masih memerlukan database uji terpisah.
