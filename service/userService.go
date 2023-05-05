package service

import (
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/mongodb"
	"myGinFrame/mysql"
)

type UserServie interface {
	NewUser(userName, tel, gender string) error
	GetUser(userId string) *model.User
	DeleteUser(userId string) error
	UpdateUser(userId, username, tel, gender string) error
}

type userService struct {
	userMysqlDao mysql.UserDao
	userMongoDao mongodb.BaseDao
}

func NewUserService() UserServie {
	return &userService{
		userMysqlDao: mysql.NewUserDaoManage(),
		userMongoDao: mongodb.NewBaseDaoManage("numbers"),
	}
}

func (s *userService) NewUser(userName, tel, gender string) error {
	s.userMongoDao.Create(&model.User{
		Name:   userName,
		Tel:    tel,
		Gender: 0})
	return s.userMysqlDao.Create(&model.User{
		Name:   userName,
		Tel:    tel,
		Gender: 0,
		//CreditCards: []*model.CreditCard{{CardNumber: "65535"}, {CardNumber: "10086"}},
		//Languages:   []*model.Language{{Name: "English"}, {Name: "Franch"}},
	})
}

func (s *userService) GetUser(userId string) *model.User {
	//s.userMongoDao.UpdateDeleteBy(map[string]interface{}{"name": "ls"}, []string{"tags"}, true)
	//glog.Glog.Info("count:", s.userMongoDao.GetCountBy(map[string]interface{}{}))

	//s.userMongoDao.Watch()
	return &model.User{Name: "zhangsan", Tel: "13888888888", Gender: 1}
}

func (s *userService) DeleteUser(userId string) error {
	glog.Glog.Info("del userId:", userId)
	return nil
}

func (s *userService) UpdateUser(userId, username, tel, gender string) error {
	glog.Glog.Info("update userId:", userId, username, tel, gender)
	return nil
}
