package mongodb

import (
	"context"
	"myGinFrame/glog"
)

type NumberDao interface {
	BaseDao
}

type numberDao struct {
	BaseDaoManage
}

func NewNumberDao() NumberDao {
	return &numberDao{BaseDaoManage: BaseDaoManage{
		ctx:       context.Background(),
		tableName: "numbers",
		conn:      w_db,
	}}
}

func (d *numberDao) GetSumByGroupKey(results interface{}, where map[string]interface{}, groupByKey string) error {
	glog.Glog.Info("调用子类")
	return nil
}
