package mongodb

import (
	"encoding/json"
	"fmt"
	"myGinFrame/tool"
	"strings"
)

//echo 0123 | sudo -S docker exec mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json'
//docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 --authenticationDatabase admin -d gin_test -c numbers -o /home/numbers.json'
//docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 --authenticationDatabase admin -d gin_test -c numbers -q {\""name\"":\""pi\""} -o /home/numbers.json'
//docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 --authenticationDatabase admin -d gin_test -c numbers -q {\""value\"":{\""\$gt\"":3.17}} -o /home/numbers.json'
func ExportMongoTableToJson(dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName, savePath string, query map[string]interface{}) error {
	//在执行docker命令时，省掉-it，将/bin/bash换成sh
	//cmdStr := "docker exec mongo-server sh -c \"mongod --version\""
	//cmdStr := "docker exec mongo-server sh -c \"mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json\""
	cmdStr := fmt.Sprintf("docker exec %s sh -c 'mongoexport -h %s -u %s -p %s --authenticationDatabase admin -d %s -c %s", dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName)
	if len(query) > 0 {
		queryStr := mapToJsonStr(query)
		//转义字符
		queryStr = strings.ReplaceAll(queryStr, "\"", "\\\"")
		queryStr = strings.ReplaceAll(queryStr, "$", "\\$")
		cmdStr += fmt.Sprintf(" -q %s", queryStr)
	}
	cmdStr += fmt.Sprintf(" -o %s'", savePath)
	result, err := tool.RunShell(cmdStr)
	tool.Info.Println("result:", result, "->err:", err)
	return err
}

//docker exec -it mongo-server /bin/bash -c 'mongoimport -h 127.0.0.1:27017 -u root -p 123456 --authenticationDatabase admin -d gin_test -c numbers --file /home/numbers.json'
// isCover: 是否删除之前的旧数据
func ImportJsonToMongoTable(dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName, jsonPath string, isCover bool) error {
	//cmdStr := "docker exec mongo-server sh -c \"mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json\""
	cmdStr := fmt.Sprintf("docker exec %s sh -c 'mongoimport -h %s -u %s -p %s --authenticationDatabase admin -d %s -c %s", dockerImageName, mongoAddr, mongoUser, mongoPwd, databaseName, tableName)
	if isCover {
		cmdStr += " --drop"
	}
	cmdStr += fmt.Sprintf(" --file %s'", jsonPath)
	result, err := tool.RunShell(cmdStr)
	tool.Info.Println("result:", result, "->err:", err)
	return err
}

func mapToJsonStr(param map[string]interface{}) string {
	dataType, _ := json.Marshal(param)
	dataString := string(dataType)
	return dataString
}

func jsonStrToMap(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		return nil
	}
	return tempMap
}
