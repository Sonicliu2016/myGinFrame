package mysql

import (
	"myGinFrame/model"
	"sync"
)

type UserDao interface {
	BaseDao
}

var userDaoManage *UserDaoManage

type UserDaoManage struct {
	BaseDaoManage
}

func NewUserDaoManage() UserDao {
	var once sync.Once
	once.Do(func() {
		userDaoManage = &UserDaoManage{BaseDaoManage{tableName: new(model.User).TableName(), mysqlConn: w_db}}
	})
	return userDaoManage
}
