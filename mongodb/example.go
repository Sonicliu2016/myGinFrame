package mongodb

import (
	"myGinFrame/glog"
)

var dao BaseDao

func initExample() {
	glog.Glog.Info("mongo test init!")
	dao = NewNumberDao()
	//dao = NewBaseDaoManage("numbers")
	//testExportMongoTable()
	//testImportToMongo()
	//testCreate()
	//testDelete()
	//testUpdateInc()
	//testUpdatePull()
	//testUpdatePush()
	//testUpdateDelete()
	//testGetCount()
	//testGetDistinct()
	//testGetManyByManyBySort()
	//testGetManyLike()
	testGetSumByGroupKey()
}

func testCreate() {
	//result, err := userMongoDao.Create(map[string]interface{}{"name": "pi", "value": 3.14})
	//glog.Glog.Info("result:", result, "->err:", err)
	//{name:'zs', books:[{name:'html', price:66}, {name:'js', price:88}], tags:['html', 'js', ['1', '2']]},
	//{name:'ls', books:[{name:'vue', price:99}, {name:'node', price:199}], tags:['a', 'b', 'c', 'd', 'e']},
	datas := make([]interface{}, 0)
	datas = append(datas, map[string]interface{}{"name": "zhangsan",
		"books": []interface{}{map[string]interface{}{"bookName": "html", "price": 20}, map[string]interface{}{"bookName": "js", "price": 30}},
		"tags":  []interface{}{"html", "js", []string{"1", "2"}},
	})
	datas = append(datas, map[string]interface{}{"name": "lisi",
		"books": []interface{}{map[string]interface{}{"bookName": "java", "price": 35}, map[string]interface{}{"bookName": "Golang", "price": 40}},
		"tags":  []interface{}{"java", "Golang", []string{"3", "4"}},
	})
	datas = append(datas, map[string]interface{}{"key": "a", "value": 3.1})
	datas = append(datas, map[string]interface{}{"key": "b", "value": 3.2})
	datas = append(datas, map[string]interface{}{"key": "c", "value": 3.3})
	datas = append(datas, map[string]interface{}{"key": "d", "value": 3.4})
	datas = append(datas, map[string]interface{}{"key": "e", "value": 3.5})
	dao.CreateMany(datas)
}

func testDelete() {
	dao.DeleteBy(map[string]interface{}{"value": map[string]interface{}{"$lt": 3.5}})
}

func testUpdateInc() {
	glog.Glog.Error("err:", dao.UpdateIncBy(map[string]interface{}{"key": "a"}, map[string]interface{}{"value": 1.2}, true))
}

func testUpdatePull() {
	//删除name为zhangsan，books数组中bookName为js的元素
	dao.UpdatePullBy(map[string]interface{}{"name": "zhangsan"}, map[string]interface{}{"books": map[string]string{"bookName": "js"}}, true)
	//删除name为zhangsan，tags数组中["1", "2"]的元素
	dao.UpdatePullBy(map[string]interface{}{"name": "zhangsan"}, map[string]interface{}{"tags": []string{"1", "2"}}, true)
}

func testUpdatePush() {
	//db.getCollection('numbers').update({ "name": "lisi"},{ "$push":	{"tags.2":"5"}})
	//如果一个字段同时被多个更新操作符更新会报错
	dao.UpdatePushBy(map[string]interface{}{"name": "lisi"}, map[string]interface{}{"books": map[string]interface{}{"bookName": "c++", "price": 50} /*"tags": "c++",*/, "tags.2": "5"}, false)
}

func testUpdateDelete() {
	//删除tags
	dao.UpdateDeleteBy(map[string]interface{}{"name": "zhangsan"}, []string{"tags"}, true)
}

//数组为空
//db.getCollection("Array").find({"books":{$exists:true} ,$where: "this.books.length <= 0"})//数组length<= 0
//db.getCollection("Array").find({"books.0":{$exists:0}})//数组第一个元素不存在
//db.getCollection("Array").find({"books": []})//数组=[]
//db.getCollection("Array").find({"books":{$size:0}})//数组的size为零
//db.getCollection("Array").find({"vendor":{$not:{$elemMatch:{$ne:null}}}})//数组elemMatch是null
//数组非空
//db.getCollection('numbers').find({"books":{$exists:true} ,$where: "this.books.length >= 2"})//数组length> 0
//db.getCollection("Array").find({"books.0":{$exists:1}})//数组第一个元素存在
//db.getCollection("Array").find({"books": {$gt:[]}})//数组大于[]
//db.getCollection("Array").find({books:{$not:{$size:0}}})数组的size不为零
//db.getCollection("Array").find({"books":{$elemMatch:{$ne:null}}})//数组elemMatch不是null
func testGetCount() {
	glog.Glog.Info("count1:", dao.GetCountBy(map[string]interface{}{"books": map[string]interface{}{"$size": 2}}))
	glog.Glog.Info("count2:", dao.GetCountBy(map[string]interface{}{"books.1": map[string]interface{}{"$exists": 1}}))
}

func testGetDistinct() {
	var bookNames []string
	dao.GetDistinctBy(&bookNames, "books.bookName", map[string]interface{}{})
	glog.Glog.Info("bookNames:", bookNames)
}

func testGetManyByManyBySort() {
	var results []interface{}
	//dao.GetManyByManyBySort(&results, map[string]interface{}{"value": map[string]interface{}{"$gt": 3}}, map[string]int{"value": 1})
	dao.GetManyByManyBySortAndSkipLimit(&results, map[string]interface{}{"value": map[string]interface{}{"$gt": 3}}, map[string]int{"value": 1}, 2, 1)
	glog.Glog.Info("results:", results)
}

func testGetManyLike() {
	var results []interface{}
	dao.GetManyLike(&results, map[string]interface{}{}, map[string]string{"name": "p"})
	glog.Glog.Info("results:", results)
}

func testGetSumByGroupKey() {
	var results []interface{}
	dao.GetSumByGroupKey(&results,
		map[string]interface{}{"key": map[string]interface{}{"$exists": true}},
		"key")
	glog.Glog.Info("results:", results)
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
