package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// Value receiver karena hanya membaca data
func (s Student) GetInfo() string {
	return fmt.Sprintf(
		"ID: %d | Name: %s | Grade: %.2f | Active: %v",
		s.ID,
		s.Name,
		s.Grade,
		s.IsActive,
	)
}

// Pointer receiver karena mengubah Grade
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Pointer receiver karena mengubah IsActive
func (s *Student) Activate() {
	s.IsActive = true
}

// Pointer receiver karena mengubah IsActive
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	student := Student{
		ID:       1,
		Name:     "Cindy",
		Grade:    85.5,
		IsActive: false,
	}

	fmt.Println("=== Data Awal Student ===")
	fmt.Println(student.GetInfo())

	fmt.Println("\n=== Setelah Activate ===")
	student.Activate()
	fmt.Println(student.GetInfo())

	fmt.Println("\n=== Setelah UpdateGrade ===")
	student.UpdateGrade(92.5)
	fmt.Println(student.GetInfo())

	fmt.Println("\n=== Setelah Deactivate ===")
	student.Deactivate()
	fmt.Println(student.GetInfo())
}