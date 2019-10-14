package main

// Employee represents a user who can log time
type Employee struct {
	// ID is the primary key
	ID int `db:"id"`

	// Name is their full name
	Name string `db:"name"`
}
