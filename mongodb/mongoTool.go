package mongodb

import (
	"fmt"
	"myGinFrame/tool"
)

//echo 0123 | sudo -S docker exec mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json'
//docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json'
//docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers -q {\""name\"":\""pi\""} --authenticationDatabase admin -o /home/numbers.json'
//https://blog.csdn.net/qqqweiweiqq/article/details/128538960
func ExportMongoTable(dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName, savePath string) {
	//在执行docker命令时，省掉-it，将/bin/bash换成sh
	//cmdStr := "docker exec mongo-server sh -c \"mongod --version\""
	//cmdStr := "docker exec mongo-server sh -c \"mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json\""
	cmdStr := fmt.Sprintf("docker exec %s sh -c 'mongoexport -h %s -u %s -p %s -d %s -c %s --authenticationDatabase admin -o %s'", dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName, savePath)
	result, err := tool.RunShell(cmdStr)
	tool.Info.Println("result:", result, "->err:", err)
}
