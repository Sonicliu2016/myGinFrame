package mysql

import (
	"gorm.io/gorm"
	"myGinFrame/tool"
	"sync"
)

type BaseDao interface {
	IsRecordExist(model interface{}, filter map[string]interface{}) bool
	Create(model interface{}) error
	Count(filter map[string]interface{}) int64
	GetAll(models interface{}) error
	GetInById(models interface{}, key string, inWhere []interface{}) error
	GetOneByKey(model interface{}, key, value string) error
	GetOneByMany(model interface{}, filter map[string]interface{}) error
	GetManyByKey(models interface{}, key, value string, limit, offset int) error
	GetManyByMany(models interface{}, filter map[string]interface{}, limit, offset int) error
	UpdateBy(filter map[string]interface{}, updateFields map[string]interface{}) error
	DeleteBy(model interface{}, filter map[string]interface{}) error
}

type BaseDaoManage struct {
	tableName string
	mysqlConn *gorm.DB
	sync.RWMutex
}

var baseDaoManage *BaseDaoManage

func NewBaseDaoManage(tableName string) BaseDao {
	var once sync.Once
	once.Do(func() {
		baseDaoManage = &BaseDaoManage{tableName: tableName, mysqlConn: w_db}
	})
	return baseDaoManage
}

//func (d *BaseDaoManage) TableName() string {
//	object := reflect.ValueOf(d.model)
//	f := object.MethodByName("TableName")
//	return f.Call([]reflect.Value{})[0].String()
//}

func (d *BaseDaoManage) IsRecordExist(model interface{}, filter map[string]interface{}) bool {
	db := d.mysqlConn
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	err := db.First(model).Error
	if err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

func (d *BaseDaoManage) Create(model interface{}) error {
	return d.mysqlConn.Create(model).Error
}

func (d *BaseDaoManage) Count(filter map[string]interface{}) int64 {
	var count int64
	db := d.mysqlConn.Table(d.tableName)
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	if err := db.Count(&count).Error; err != nil {
		tool.Error.Println(d.tableName, " get count err:", err)
		return 0
	}
	return count
}

func (d *BaseDaoManage) GetAll(models interface{}) error {
	return d.mysqlConn.Find(models).Error
}

func (d *BaseDaoManage) GetInById(models interface{}, key string, inWhere []interface{}) error {
	return d.mysqlConn.Where(key+" IN ?", inWhere).Find(models).Error
}

func (d *BaseDaoManage) GetOneByKey(model interface{}, key, value string) error {
	return d.mysqlConn.Where(key+" = ?", value).First(model).Error
}

func (d *BaseDaoManage) GetOneByMany(model interface{}, filter map[string]interface{}) error {
	db := d.mysqlConn
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	return db.First(model).Error
}

func (d *BaseDaoManage) GetManyByKey(models interface{}, key, value string, limit, offset int) error {
	db := d.mysqlConn.Where(key+" = ?", value)
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}
	return db.Find(models).Error
}

func (d *BaseDaoManage) GetManyByMany(models interface{}, filter map[string]interface{}, limit, offset int) error {
	db := d.mysqlConn
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}
	return db.Find(models).Error
}

func (d *BaseDaoManage) UpdateBy(filter map[string]interface{}, updateFields map[string]interface{}) error {
	db := d.mysqlConn.Table(d.tableName)
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	return db.Updates(updateFields).Error
}

func (d *BaseDaoManage) DeleteBy(model interface{}, filter map[string]interface{}) error {
	db := d.mysqlConn
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	return db.Delete(model).Error
}
