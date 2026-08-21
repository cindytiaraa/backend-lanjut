package main

import "fmt"

// Menukar nilai dua integer menggunakan pointer
func swap(a, b *int) {
	*a, *b = *b, *a
}

// Menambahkan item baru ke slice menggunakan pointer
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// Pass by value
func passByValue(x int) {
	x = 100
}

// Pass by pointer
func passByPointer(x *int) {
	*x = 100
}

func main() {
	fmt.Println("=== Swap dengan Pointer ===")

	a := 10
	b := 20

	fmt.Println("Sebelum swap:", a, b)

	swap(&a, &b)

	fmt.Println("Sesudah swap:", a, b)

	fmt.Println("\n=== Update Slice dengan Pointer ===")

	items := []string{"Go", "Fiber"}

	fmt.Println("Sebelum update:", items)

	updateSlice(&items, "Backend")

	fmt.Println("Sesudah update:", items)

	fmt.Println("\n=== Pass by Value vs Pass by Pointer ===")

	value := 50

	fmt.Println("Nilai awal:", value)

	passByValue(value)

	fmt.Println("Setelah pass by value:", value)

	passByPointer(&value)

	fmt.Println("Setelah pass by pointer:", value)
}