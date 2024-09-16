package db

type Record struct {
	Base

	Name   string `json:"name"`
	Target string `json:"target"`
}
