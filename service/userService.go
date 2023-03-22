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
	mongodb.ExportMongoTable("mongo-server", "127.0.0.1:27017", "root", "123456", "gin_test", "numbers", "/home/numbers.json", nil)
	//s.userMongoDao.Watch()
	var names []float64
	s.userMongoDao.GetDistinctBy(&names, "value", map[string]interface{}{})
	glog.Glog.Info("names:", names)
	return &model.User{Name: "zhangsan", Tel: "13888888888", Gender: 1}
}

func (s *userService) DeleteUser(userId string) error {
	glog.Glog.Info("del userId:", userId)
	s.userMongoDao.UpdatePushBy(map[string]interface{}{"name": "ls"}, map[string]interface{}{"books": map[string]interface{}{"name": "golang", "price": 1000}}, true)
	return nil
}

func (s *userService) UpdateUser(userId, username, tel, gender string) error {
	glog.Glog.Info("update userId:", userId, username, tel, gender)
	s.userMongoDao.UpdatePullBy(map[string]interface{}{"name": "zs"}, map[string]interface{}{"books": map[string]string{"name": "js"}}, true)
	s.userMongoDao.UpdatePullBy(map[string]interface{}{"name": "zs"}, map[string]interface{}{"t![](../../../../../../media/xufeng8/share/F01.bmp)ags": []string{"1", "2"}}, true)
	return nil
}
