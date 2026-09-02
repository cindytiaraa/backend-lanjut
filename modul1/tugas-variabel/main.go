package main

import "fmt"

func main() {
	// Lima variabel dengan tipe berbeda
	var nama string = "Cindy"
	var umur int = 20
	var nilai float64 = 90.5
	var isActive bool = true
	var mataKuliah []string = []string{
		"Pemrograman Backend",
		"Basis Data",
		"Rekayasa Perangkat Lunak",
	}

	fmt.Println("=== Variabel ===")
	fmt.Println("Nama:", nama)
	fmt.Println("Umur:", umur)
	fmt.Println("Nilai:", nilai)
	fmt.Println("Aktif:", isActive)
	fmt.Println("Mata Kuliah:", mataKuliah)

	// Map data mahasiswa
	mahasiswa := make(map[string]int)

	// Menambah data
	mahasiswa["Cindy"] = 90
	mahasiswa["Alya"] = 85
	mahasiswa["Budi"] = 88

	fmt.Println("\n=== Data Mahasiswa ===")
	for nama, nilai := range mahasiswa {
		fmt.Println(nama, ":", nilai)
	}

	// Membaca data dengan pengecekan keberadaan
	fmt.Println("\n=== Pengecekan Data ===")

	if nilai, ada := mahasiswa["Cindy"]; ada {
		fmt.Println("Cindy ditemukan dengan nilai:", nilai)
	} else {
		fmt.Println("Cindy tidak ditemukan")
	}

	if nilai, ada := mahasiswa["Dina"]; ada {
		fmt.Println("Dina ditemukan dengan nilai:", nilai)
	} else {
		fmt.Println("Dina tidak ditemukan")
	}

	// Menghapus data
	delete(mahasiswa, "Budi")

	fmt.Println("\n=== Setelah Penghapusan Budi ===")
	for nama, nilai := range mahasiswa {
		fmt.Println(nama, ":", nilai)
	}

	// Menelusuri seluruh isi map
	fmt.Println("\n=== Seluruh Data Mahasiswa ===")
	for nama, nilai := range mahasiswa {
		fmt.Println("Nama:", nama, "| Nilai:", nilai)
	}
}