package models

import "time"

type Url struct{
	ID int
	Original string
	ShortCode string
	CreatedAt time.Time
	LastAccessed time.Time
	Clicks int
}

