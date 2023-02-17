package service

import (
	"myGinFrame/glog"
	"myGinFrame/model"
)

type UserServie interface {
	NewUser(userName, tel, gender string) error
	GetUser(userId string) *model.User
	DeleteUser(userId string) error
	UpdateUser(userId, username, tel, gender string) error
}

type userService struct {
}

func NewUserService() UserServie {
	return &userService{}
}

func (s *userService) NewUser(userName, tel, gender string) error {
	glog.Glog.Info("new user:", userName, tel, gender)
	return nil
}

func (s *userService) GetUser(userId string) *model.User {
	return &model.User{UserName: "zhangsan", Tel: "13888888888", Gender: 1}
}

func (s *userService) DeleteUser(userId string) error {
	glog.Glog.Info("del userId:", userId)
	return nil
}

func (s *userService) UpdateUser(userId, username, tel, gender string) error {
	glog.Glog.Info("update userId:", userId, username, tel, gender)
	return nil
}
