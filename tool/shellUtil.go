package tool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
//RunShell("sshpass","-p",pwd,"rsync","-ra",srcpath,destpath)
//RunShell("ffmpeg","-i",videoPath,"-y","-f","image2","-t","0.001",destpath) -> ffmpeg -i /home/1.mp4 -y -f image2 -t 0.001 /home/out.jpg
func RunShell(s string, arg ...string) (string, error) {
	cmd := exec.Command(s, arg...)
	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	fmt.Println("cmd:",cmd.String())
	err := cmd.Run()
	return out.String(), err
}

/**
 * 备份MySql数据库
 * @param 	host: 			数据库地址: localhost
 * @param 	port:			端口: 3306
 * @param 	user:			用户名: root
 * @param 	password:		密码: root
 * @param 	databaseName:	需要被分的数据库名: test
 * @param 	tableName:		需要备份的表名: user
 * @param 	sqlPath:		备份SQL存储路径: D:/backup/test/
 * @return 	backupPath
 */
func BackupMySqlDb(host, port, user, password, databaseName, tableName, sqlPath string) (error, string) {
	var cmd *exec.Cmd
	if tableName == "" {
		cmd = exec.Command("mysqldump", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName)
	} else {
		cmd = exec.Command("mysqldump", "--opt", "-h"+host, "-P"+port, "-u"+user, "-p"+password, databaseName, tableName)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return err, ""
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return err, ""
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
		return err, ""
	}
	//now := time.Now().Format("2006_0102_1504")
	var backupPath string
	if tableName == "" {
		backupPath = sqlPath + "/" + databaseName + ".sql"
	} else {
		backupPath = sqlPath + "/" + databaseName + "_" + tableName + ".sql"
	}
	err = ioutil.WriteFile(backupPath, bytes, 0644)

	if err != nil {
		panic(err)
		return err, ""
	}
	return nil, backupPath
}
