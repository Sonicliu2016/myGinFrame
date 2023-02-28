package model

type User struct {
	Id       int    `json:"id"`
	UserName string `json:"userName"`
	Tel      string `json:"tel"`
	Gender   int    `json:"gender"`
}
