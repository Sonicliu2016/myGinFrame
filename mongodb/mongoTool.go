package mongodb

import "myGinFrame/tool"

//echo 0123 | sudo -S docker exec mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json'
//sudo docker exec -it mongo-server /bin/bash -c 'mongoexport -h 127.0.0.1:27017 -u root -p 123456 -d gin_test -c numbers --authenticationDatabase admin -o /home/numbers.json'
func ExportMongoTable()  {
	params := []string{"exec","mongo-server","/bin/bash","-c","\"","mongoexport" ,"-h" ,"127.0.0.1:27017","-u", "root", "-p", "123456", "-d", "gin_test", "-c", "numbers" ,"--authenticationDatabase", "admin", "-o" ,"/home/numbers.json","\""}
	r,err := tool.RunShell("docker",params...)
	tool.Info.Println("r:",r,"->err:",err)
}
