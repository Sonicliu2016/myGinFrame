package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"myGinFrame/tool"
	"sync"
	"time"
)

type BaseDao interface {
	coll() *mongo.Collection
	Create(one interface{}) (*mongo.InsertOneResult, error)
	CreateMany(many []interface{}) (*mongo.InsertManyResult, error)
	Delete(id string) error
	DeleteBy(where map[string]interface{}) error

	Update(id string, updateFields map[string]interface{}) error
	UpdateBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//自增或自减文档中的某个int值
	UpdateIncBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	//删除文档中数组中的元素
	UpdatePullBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error
	UpdatePushBy(where map[string]interface{}, updateFields map[string]interface{}, updateMany bool) error

	Get(v interface{}, id string) error
	GetBy(m interface{}, key, value string) error
	GetManyBy(m interface{}, key, value string) error
	GetManyByMany(m interface{}, filter map[string]interface{}) error
	GetManyByManyBySort(m interface{}, filter map[string]interface{}, sortBy map[string]interface{}) error
	GetManyIn(m interface{}, key string, values []interface{}) error
	IsRecordExist(filter map[string]interface{}) bool
	GetOneOrder(m interface{}, key string, order int) error
	GetOneByMany(m interface{}, filter map[string]interface{}) error
	GetManyLike(m interface{}, filter map[string]interface{}, likes map[string]string) error
	All(m interface{}) error
}

//bson.D是BSON文档的有序表示 bson.D{{"foo", "bar"}, {"hello", "world"}, {"pi", 3.14159}}
//bson.M是BSON文档的无序表示 bson.M{"foo": "bar", "hello": "world", "pi": 3.14159}
//bson.A是BSON数组的有序表示 bson.A{"bar", "world", 3.14159, bson.D{{"qux", 12345}}}
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

func (d *BaseDaoManage) Create(one interface{}) (*mongo.InsertOneResult, error) {
	return d.coll().InsertOne(d.ctx, one)
}

func (d *BaseDaoManage) CreateMany(many []interface{}) (*mongo.InsertManyResult, error) {
	return d.coll().InsertMany(d.ctx, many)
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

func (d *BaseDaoManage) UpdateDeleteBy() {

}

func (d *BaseDaoManage) IsRecordExist(filter map[string]interface{}) bool {
	f := bson.M{}
	for k, v := range filter {
		f[k] = v
	}
	cur, err := d.coll().Find(d.ctx, f)
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

func (d *BaseDaoManage) GetOneOrder(m interface{}, key string, order int) error {
	opts := options.FindOne()
	opts.SetSort(bson.D{{Key: key, Value: order}})
	return d.coll().FindOne(d.ctx, bson.M{}, opts).Decode(m)
}

func (d *BaseDaoManage) Get(m interface{}, id string) error {
	return d.coll().FindOne(d.ctx, bson.M{"model.id": id}).Decode(m)
}

func (d *BaseDaoManage) GetBy(m interface{}, key, value string) error {
	return d.coll().FindOne(d.ctx, bson.M{key: value}).Decode(m)
}

func (d *BaseDaoManage) GetManyBy(m interface{}, key, value string) error {
	c, e := d.coll().Find(d.ctx, bson.M{key: value})
	if e != nil {
		return e
	}
	return c.All(d.ctx, m)
}

func (d *BaseDaoManage) GetManyByMany(m interface{}, filter map[string]interface{}) error {
	f := bson.M{"model.deletedAt": nil}
	for k, v := range filter {
		f[k] = v
	}
	c, e := d.coll().Find(d.ctx, f)
	if e != nil {
		return e
	}
	return c.All(d.ctx, m)
}

// 其中 1 为升序排列，而 -1 是用于降序排列
func (d *BaseDaoManage) GetManyByManyBySort(m interface{}, filter map[string]interface{}, sortBy map[string]interface{}) error {
	f := bson.M{"model.deletedAt": nil}
	for k, v := range filter {
		f[k] = v
	}
	sort := bson.D{}
	for k, v := range sortBy {
		sort = append(sort, bson.E{Key: k, Value: v})
	}
	c, e := d.coll().Aggregate(d.ctx, bson.A{
		bson.M{
			"$sort": sort,
		},
		bson.M{
			"$match": f,
		},
	})
	if e != nil {
		return e
	}
	return c.All(d.ctx, m)
}

func (d *BaseDaoManage) GetManyIn(m interface{}, key string, values []interface{}) error {
	f := bson.M{key: bson.M{"$in": values}}
	c, e := d.coll().Find(d.ctx, f)
	if e != nil {
		return e
	}
	return c.All(d.ctx, m)
}

func (d *BaseDaoManage) GetOneByMany(m interface{}, filter map[string]interface{}) error {
	f := bson.M{"model.deletedAt": nil}
	for k, v := range filter {
		f[k] = v
	}
	return d.coll().FindOne(d.ctx, f).Decode(m)
}

func (d *BaseDaoManage) GetManyLike(m interface{}, filter map[string]interface{}, likes map[string]string) error {
	f := bson.M{"model.deletedAt": nil}
	for k, v := range filter {
		f[k] = v
	}
	for k, v := range likes {
		f[k] = primitive.Regex{Pattern: v, Options: "im"}
	}
	return d.coll().FindOne(d.ctx, f).Decode(m)
}

func (d *BaseDaoManage) All(m interface{}) error {
	c, e := d.coll().Find(d.ctx, bson.M{"model.deletedAt": nil})
	if e != nil {
		return e
	}
	return c.All(d.ctx, m)
}
