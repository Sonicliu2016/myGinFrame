package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"myGinFrame/tool"
	"reflect"
	"sync"
	"time"
)

type BaseDao interface {
	coll() *mongo.Collection
	Create(result interface{}) (*mongo.InsertOneResult, error)
	CreateMany(results []interface{}) (*mongo.InsertManyResult, error)
	Delete(id string) error
	DeleteBy(where map[string]interface{}) error
	Update(id string, updateFields map[string]interface{}) error
	UpdateBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//自增或自减文档中的某个int值
	UpdateIncBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//删除文档中数组中的元素
	UpdatePullBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//对Array(list) 数据进行增加新元素
	UpdatePushBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//删除记录中指定的key(field)
	UpdateDeleteBy(where map[string]interface{}, deleteKeys []string, updateMany bool) error

	IsRecordExist(where map[string]interface{}) bool
	Get(result interface{}, id string) error
	GetBy(result interface{}, key, value string) error
	GetOneByMany(result interface{}, where map[string]interface{}) error
	GetManyBy(results interface{}, key, value string) error
	GetManyByMany(results interface{}, where map[string]interface{}) error
	GetCountBy(where map[string]interface{}) int64
	//检索字段的非重复值
	GetDistinctBy(results interface{}, fieldName string, where map[string]interface{}) (err error)
	GetOneOrder(result interface{}, key string, order int) error
	GetManyByManyBySort(results interface{}, where map[string]interface{}, sortBy map[string]int) error
	GetManyIn(results interface{}, key string, values []interface{}) error
	GetManyLike(results interface{}, where map[string]interface{}, likes map[string]string) error
	All(results interface{}) error
	//有问题
	Watch()
}

//bson.D是BSON文档的有序表示 bson.D{{"foo", "bar"}, {"hello", "world"}, {"pi", 3.14159}}
//bson.M是BSON文档的无序表示 bson.M{"foo": "bar", "hello": "world", "pi": 3.14159}
//bson.A是BSON数组的有序表示 bson.A{"bar", "world", 3.14159, bson.D{{"qux", 12345}}}
//https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/query-document/
type BaseDaoManage struct {
	ctx       context.Context
	conn      *mongo.Database
	tableName string
	sync.RWMutex
}

var baseDaoManage *BaseDaoManage

func NewBaseDaoManage(tableName string) BaseDao {
	var once sync.Once
	once.Do(func() {
		baseDaoManage = &BaseDaoManage{ctx: context.Background(), conn: w_db, tableName: tableName}
	})
	return baseDaoManage
}

func (d *BaseDaoManage) coll() *mongo.Collection {
	return d.conn.Collection(d.tableName)
}

func (d *BaseDaoManage) Create(result interface{}) (*mongo.InsertOneResult, error) {
	return d.coll().InsertOne(d.ctx, result)
}

func (d *BaseDaoManage) CreateMany(results []interface{}) (*mongo.InsertManyResult, error) {
	return d.coll().InsertMany(d.ctx, results)
}

func (d *BaseDaoManage) Delete(id string) error {
	_, err := d.coll().DeleteOne(d.ctx, bson.M{"model.id": id})
	return err
}

func (d *BaseDaoManage) DeleteBy(where map[string]interface{}) error {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	_, err := d.coll().DeleteMany(d.ctx, filter)
	return err
}

func (d *BaseDaoManage) Update(id string, updateFields map[string]interface{}) error {
	updateFields["model.updatedAt"] = time.Now()
	_, err := d.coll().UpdateOne(d.ctx, bson.M{"model.id": id}, bson.M{"$set": updateFields})
	return err
}

func (d *BaseDaoManage) UpdateBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error {
	updateFields["model.updatedAt"] = time.Now()
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	if updateMany {
		_, err := d.coll().UpdateMany(d.ctx, filter, bson.M{"$set": updateFields})
		return err
	} else {
		_, err := d.coll().UpdateOne(d.ctx, filter, bson.M{"$set": updateFields})
		return err
	}
}

//自增或自减文档中的某个int值
func (d *BaseDaoManage) UpdateIncBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error {
	updateFields["model.updatedAt"] = time.Now()
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	if updateMany {
		_, err := d.coll().UpdateMany(d.ctx, filter, bson.M{"$inc": updateFields})
		return err
	} else {
		_, err := d.coll().UpdateOne(d.ctx, filter, bson.M{"$inc": updateFields})
		return err
	}
}

//指定删除Array中的某一个元素
//db.getCollection('numbers').insert([
//{name:'zs', books:[{name:'html', price:66}, {name:'js', price:88}], tags:['html', 'js', ['1', '2']]},
//{name:'ls', books:[{name:'vue', price:99}, {name:'node', price:199}], tags:['a', 'b', 'c', 'd', 'e']},
//])
//删除name为zs，book中name为js的值
//db.getCollection('numbers').updateOne({name:'zs'},{$pull:{books:{name:'js'}}})
//如果要删除的元素是一个数组, 那么必须一模一样才能删除
//db.person.updateOne({name:'zs'},{$pull:{tags:['1','2']}})
//s.userMongoDao.UpdatePullBy(map[string]interface{}{"name": "zs"}, map[string]interface{}{"books": map[string]string{"name": "js"}}, true)
//s.userMongoDao.UpdatePullBy(map[string]interface{}{"name": "zs"}, map[string]interface{}{"tags": []string{"1", "2"}}, true)
func (d *BaseDaoManage) UpdatePullBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	if updateMany {
		_, err := d.coll().UpdateMany(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$pull": updateFields})
		return err
	} else {
		_, err := d.coll().UpdateOne(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$pull": updateFields})
		return err
	}
}

//对Array(list)数据进行增加新元素
//s.userMongoDao.UpdatePushBy(map[string]interface{}{"name": "ls"}, map[string]interface{}{"books": map[string]interface{}{"name": "golang", "price": 1000}}, true)
func (d *BaseDaoManage) UpdatePushBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error {
	filter := bson.D{}
	for k, v := range where {
		filter = append(filter, bson.E{Key: k, Value: v})
	}
	if updateMany {
		_, err := d.coll().UpdateMany(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$push": updateFields})
		return err
	} else {
		_, err := d.coll().UpdateOne(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$push": updateFields})
		return err
	}
}

//删除记录中指定的key(field)
//s.userMongoDao.UpdateDeleteBy(map[string]interface{}{"name": "ls"}, []string{"tags"}, true)
func (d *BaseDaoManage) UpdateDeleteBy(where map[string]interface{}, deleteKeys []string, updateMany bool) error {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	updateFields := make(map[string]bool)
	for _, key := range deleteKeys {
		updateFields[key] = true
	}
	if updateMany {
		_, err := d.coll().UpdateMany(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$unset": updateFields})
		return err
	} else {
		_, err := d.coll().UpdateOne(d.ctx, filter, bson.M{"$set": bson.M{"model.updatedAt": time.Now()}, "$unset": updateFields})
		return err
	}
}

func (d *BaseDaoManage) IsRecordExist(where map[string]interface{}) bool {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	cur, err := d.coll().Find(d.ctx, filter)
	if err != nil {
		tool.Error.Println("Find err:", err)
		return false
	}
	var m []interface{}
	if err = cur.All(d.ctx, &m); err != nil {
		tool.Error.Println("cur.All err:", err)
		return false
	}
	if len(m) > 0 {
		return true
	}
	return false
}

func (d *BaseDaoManage) Get(result interface{}, id string) error {
	return d.coll().FindOne(d.ctx, bson.M{"model.id": id}).Decode(result)
}

func (d *BaseDaoManage) GetBy(result interface{}, key, value string) error {
	return d.coll().FindOne(d.ctx, bson.M{key: value}).Decode(result)
}

func (d *BaseDaoManage) GetOneByMany(result interface{}, where map[string]interface{}) error {
	filter := bson.M{"model.deletedAt": nil}
	for k, v := range where {
		filter[k] = v
	}
	return d.coll().FindOne(d.ctx, filter).Decode(result)
}

func (d *BaseDaoManage) GetManyBy(results interface{}, key, value string) error {
	cursor, err := d.coll().Find(d.ctx, bson.M{key: value})
	if err != nil {
		return err
	}
	return cursor.All(d.ctx, results)
}

func (d *BaseDaoManage) GetManyByMany(results interface{}, where map[string]interface{}) error {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	cursor, err := d.coll().Find(d.ctx, filter)
	if err != nil {
		return err
	}
	return cursor.All(d.ctx, results)
}

func (d *BaseDaoManage) GetCountBy(where map[string]interface{}) int64 {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	count, err := d.coll().CountDocuments(d.ctx, filter)
	if err != nil {
		tool.Error.Println("CountDocuments err:", err)
		return 0
	}
	return count
}

//获取去除重复后字段的值
//db.getCollection('numbers').distinct("books.name")
//s.userMongoDao.GetDistinctBy(&names, "books.name", map[string]interface{}{})
func (d *BaseDaoManage) GetDistinctBy(results interface{}, fieldName string, where map[string]interface{}) error {
	filter := bson.M{}
	for k, v := range where {
		filter[k] = v
	}
	rs, err := d.coll().Distinct(d.ctx, fieldName, filter)

	reflectResultsVal := reflect.ValueOf(results)
	if reflectResultsVal.Kind() != reflect.Ptr {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a %s", reflectResultsVal.Kind())
	}
	sliceVal := reflectResultsVal.Elem()
	if sliceVal.Kind() == reflect.Interface {
		sliceVal = sliceVal.Elem()
	}
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a pointer to %s", sliceVal.Kind())
	}
	for _, r := range rs {
		s := reflect.ValueOf(r)
		newElem := reflect.New(s.Type()).Elem()
		newElem.Set(s)
		sliceVal = reflect.Append(sliceVal, newElem)
	}
	reflectResultsVal.Elem().Set(sliceVal)
	return err
}

func (d *BaseDaoManage) GetOneOrder(result interface{}, key string, order int) error {
	opts := options.FindOne()
	opts.SetSort(bson.D{{Key: key, Value: order}})
	return d.coll().FindOne(d.ctx, bson.M{}, opts).Decode(result)
}

// sortBy:其中 1 为升序排列，而 -1 是用于降序排列
func (d *BaseDaoManage) GetManyByManyBySort(results interface{}, where map[string]interface{}, sortBy map[string]int) error {
	filter := bson.M{"model.deletedAt": nil}
	for k, v := range where {
		filter[k] = v
	}
	sort := bson.D{}
	for k, v := range sortBy {
		sort = append(sort, bson.E{Key: k, Value: v})
	}
	cursor, err := d.coll().Aggregate(d.ctx, bson.A{
		bson.M{
			"$sort": sort,
		},
		bson.M{
			"$match": filter,
		},
	})
	if err != nil {
		return err
	}
	return cursor.All(d.ctx, results)
}

func (d *BaseDaoManage) GetManyIn(results interface{}, key string, values []interface{}) error {
	filter := bson.M{key: bson.M{"$in": values}}
	cursor, err := d.coll().Find(d.ctx, filter)
	if err != nil {
		return err
	}
	return cursor.All(d.ctx, results)
}

func (d *BaseDaoManage) GetManyLike(results interface{}, where map[string]interface{}, likes map[string]string) error {
	filter := bson.M{"model.deletedAt": nil}
	for k, v := range where {
		filter[k] = v
	}
	for k, v := range likes {
		filter[k] = primitive.Regex{Pattern: v, Options: "im"}
	}
	return d.coll().FindOne(d.ctx, filter).Decode(results)
}

func (d *BaseDaoManage) All(results interface{}) error {
	cursor, err := d.coll().Find(d.ctx, bson.M{"model.deletedAt": nil})
	if err != nil {
		return err
	}
	return cursor.All(d.ctx, results)
}

func (d *BaseDaoManage) Watch() {
	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{{"operationType", "insert"}}}}}
	cs, err := d.coll().Watch(d.ctx, pipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		tool.Error.Println("err:", err)
		return
	}
	tool.Info.Println("Waiting For Change Events. Insert something in MongoDB!")
	for cs.Next(d.ctx) {
		var event bson.M
		if err = cs.Decode(&event); err != nil {
			tool.Error.Println("err:", err)
		}
		output, err := json.MarshalIndent(event["fullDocument"], "", "    ")
		if err != nil {
			tool.Error.Println("err:", err)
		}
		tool.Info.Println("%s\n", output)
	}
	if err = cs.Err(); err != nil {
		tool.Error.Println("err:", err)
	}
}
