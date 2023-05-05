package model

import "time"

type NError struct {
	Code int
	Msg  string
}

func (e NError) Error() string {
	return e.Msg
}

type BaseInt64Model struct {
	Id        int64     `gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BaseStringModel struct {
	Id        string    `gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
