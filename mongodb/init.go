package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/tool"
	"time"
)

var w_db *mongo.Database
var ctx context.Context

func init() {
	userName := tool.GetConfigStr("mongo_w_user")
	pwd := tool.GetConfigStr("mongo_w_pwd")
	host := tool.GetConfigStr("mongo_w_host")
	port := tool.GetConfigStr("mongo_w_port")
	dbName := tool.GetConfigStr("mongo_w_database")
	w_db, _ = connectMongodb(userName, pwd, host, port, dbName)
	addIndexToMongodb(new(model.User).TableName(), []string{"name"}, []string{"tel"})
	initExample()
}

func connectMongodb(user, pwd, host, port, dbName string) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s", host, port)
	credential := options.Credential{
		Username: user,
		Password: pwd,
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(uri).SetAuth(credential)
	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		glog.Glog.Error("连接mongo数据库错误：", err.Error())
		return nil, err
	}
	// 检查连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		glog.Glog.Error("ping mongodb错误：", err.Error())
		return nil, err
	}
	db := client.Database(dbName)
	glog.Glog.Info("连接mongo数据库:", dbName, "成功！")
	//collection := db.Collection("numbers")
	//result, _ := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	//glog.Glog.Info("插入数据成功，result:", result)
	return db, nil
}

func addIndexToMongodb(table string, uniqueFields []string, fields []string) {
	collection := w_db.Collection(table)
	uniqueIndexs := make([]mongo.IndexModel, 0)
	indexs := make([]mongo.IndexModel, 0)
	//1 表示按升序创建索引，-1 表示按降序来创建索引
	uniqueIndexs = append(uniqueIndexs, mongo.IndexModel{
		Keys:    bson.D{{"model.id", 1}},
		Options: options.Index().SetUnique(true),
	})
	indexs = append(indexs, mongo.IndexModel{
		Keys: bson.D{{"model.createdAt", 1}},
	})
	indexs = append(indexs, mongo.IndexModel{
		Keys: bson.D{{"model.updatedAt", 1}},
	})
	indexs = append(indexs, mongo.IndexModel{
		Keys: bson.D{{"model.deletedAt", 1}},
	})
	for _, f := range uniqueFields {
		uniqueIndexs = append(uniqueIndexs, mongo.IndexModel{
			Keys:    bson.D{{f, 1}},
			Options: options.Index().SetUnique(true),
		})
	}
	for _, f := range fields {
		indexs = append(indexs, mongo.IndexModel{
			Keys: bson.D{{f, 1}},
		})
	}
	for _, i := range uniqueIndexs {
		name, err := collection.Indexes().CreateOne(context.Background(), i)
		if err != nil {
			glog.Glog.Error("createUniqueIndex:", err, "->table:", table, "->index:", i.Keys)
		} else {
			glog.Glog.Info("table:", table, "->Unique Index Created: "+name)
		}
	}
	for _, i := range indexs {
		name, err := collection.Indexes().CreateOne(context.Background(), i)
		if err != nil {
			glog.Glog.Error("createIndex:", err, "->table:", table, "->index:", i.Keys)
		} else {
			glog.Glog.Info("table:", table, "->Index Created: "+name)
		}
	}
}
