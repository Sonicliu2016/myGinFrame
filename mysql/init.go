package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"myGinFrame/glog"
	"myGinFrame/model"
	"myGinFrame/tool"
	"os"
	"time"
)

var w_db *gorm.DB

func init() {
	userName := tool.GetConfigStr("mysql_w_user")
	pwd := tool.GetConfigStr("mysql_w_pwd")
	host := tool.GetConfigStr("mysql_w_host")
	port := tool.GetConfigStr("mysql_w_port")
	dbName := tool.GetConfigStr("mysql_w_database")
	logLevel := tool.GetConfigInt("log_level")
	w_db = connectMysql(userName, pwd, host, port, dbName, logLevel)
}

func connectMysql(userName, pwd, host, port, dbName string, logLevel int) *gorm.DB {
	var db *gorm.DB
	var err error
	mysqlArgs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", userName, pwd, host, port, dbName)
	glog.Glog.Info("mysqlArgs:", mysqlArgs)
	loggerConfig := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second * 10,          // 慢 SQL 阈值
			LogLevel:                  logger.LogLevel(logLevel), // Log level
			IgnoreRecordNotFoundError: false,                     // 忽略ErrRecordNotFound错误
			Colorful:                  true,                      // 禁用彩色打印
		},
	)
	mysqlConfig := mysql.Config{
		DSN:                       mysqlArgs,
		SkipInitializeWithVersion: false, // 根据版本自动配置
		DefaultStringSize:         255,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
	}
	for i := 0; i < 3; i++ {
		db, err = gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
			SkipDefaultTransaction:                   true,  //禁用默认事务
			PrepareStmt:                              true,  //sql预编译
			DisableForeignKeyConstraintWhenMigrating: false, //禁用自动创建外键约束
			Logger:                                   loggerConfig,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, //禁用表名复数形式
			},
		})
		if err != nil {
			glog.Glog.Error("连接mysql数据库"+dbName+"错误：", err.Error())
			time.Sleep(time.Second * 5)
		} else {
			glog.Glog.Info("连接数据库" + dbName + "表succ")
			break
		}
	}
	sqlDb, err := db.DB()
	if err == nil {
		//设置了连接可复用的最大时间
		sqlDb.SetConnMaxLifetime(10 * time.Minute)
		//设置数据库最大空闲连接
		sqlDb.SetMaxIdleConns(30)
		//设置数据库的最大数据库连接
		sqlDb.SetMaxOpenConns(100)
	}
	db.Set("gorm:table_options", "charset=utf8").AutoMigrate(
		&model.User{},
		&model.CreditCard{},
		&model.Language{},
		&model.Folder{},
		&model.Image{},
	)
	return db
}
