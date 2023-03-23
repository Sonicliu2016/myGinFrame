package mongodb

import (
	"myGinFrame/glog"
)

var userMongoDao BaseDao

func initExample() {
	glog.Glog.Info("mongo test init!")
	userMongoDao = NewBaseDaoManage("numbers")
	//testExportMongoTable()
	//testImportToMongo()
	testCreate()
}

func testCreate() {
	result, err := userMongoDao.Create(map[string]interface{}{"name": "pi", "value": 3.14})
	glog.Glog.Info("result:", result, "->err:", err)
	data := make([]interface{}, 0)
	data = append(data, map[string]interface{}{})
	userMongoDao.CreateMany(data)
}

func testExportMongoTable() {
	query := map[string]interface{}{
		"value": map[string]interface{}{"$gt": 3.17},
	}
	ExportMongoTableToJson("mongo-server", "127.0.0.1:27017", "root", "123456", "gin_test", "numbers", "/home/numbers.json", query)
}

func testImportToMongo() {
	ImportJsonToMongoTable("mongo-server", "127.0.0.1:27017", "root", "123456", "gin_test", "numbers", "/home/numbers.json", true)
}
