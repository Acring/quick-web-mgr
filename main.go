package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"database/sql"
	_ "github.com/lib/pq"
	"errors"
)
type USER struct{  //用户信息
	Username string `json:"username"`
	Password string `json:"password"`
}

type FeedBack struct{  // 数据反馈
	Code int `json:"code"`
	Ok bool `json:"ok"`
}

const(
	host = "localhost"
	port = 5432
	user = "postgres"
	sqlpassword = "Alf1netC"
	dbname = "testdb"
)

func checkExist(username string)(bool, error){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, sqlpassword, dbname)

	db,err := sql.Open("postgres", psqlInfo)

	if err != nil{
		fmt.Print("连接数据库错误")
		return false, errors.New("连接数据库出错")
	}

	sqlStat := "SELECT from userInfo WHERE username = $1"

	stmt, err := db.Prepare(sqlStat)

	if err != nil {
		fmt.Print("准备语句失败", err.Error())
		return false, errors.New("准备语句失败")
	}

	rows ,err := stmt.Query(username)

	if rows.Next(){
		return true, nil
	}

	return false, nil

}
func checkLogin(username string, password string) (bool, error){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, sqlpassword, dbname)

	db,err := sql.Open("postgres", psqlInfo)

	if err != nil{
		fmt.Print("连接数据库错误")
		return false, errors.New("连接数据库出错")
	}

	sqlStatement := "SELECT * FROM userinfo WHERE username = $1 AND password = $2;"

	stmt, err := db.Prepare(sqlStatement)

	if err != nil {
		print(err.Error())
		return false, errors.New("准备数据失败")
	}

	rows, err := stmt.Query(username, password)
	defer rows.Close()
	var user USER
	for rows.Next(){
		rows.Scan(user.Username, user.Password)
		return true, nil
	}

	return false, nil  // 没有查询到数据
}

func checkRegister(username string , password string)(bool, error){  // 注册
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, sqlpassword, dbname)

	db,err := sql.Open("postgres", psqlInfo)

	if err != nil{
		fmt.Print("连接数据库错误")
		return false, errors.New("连接数据库出错")
	}

	sqlStatement := "INSERT INTO userinfo(username, password) VALUES ($1, $2);"

	stmt, err := db.Prepare(sqlStatement)

	if err != nil{
		fmt.Print("准备语句失败", err.Error())
		return false, errors.New("准备语句失败")
	}

	_, err = stmt.Exec(username, password)
	defer stmt.Close()
	if err != nil{
		fmt.Print("插入数据失败", err.Error())
		return false, errors.New("插入数据失败")
	}

	return true, nil

}

func loginHandle(res http.ResponseWriter, req *http.Request){ // 登录处理函数
	res.Header().Set("Access-Control-Allow-Origin", "*") //设置跨域

	result, err := ioutil.ReadAll(req.Body)
	if err != nil{
		fmt.Print(err)
		return
	}

	var userInfo USER
	err = json.Unmarshal(result, &userInfo)  // 解析json数据

	if err != nil{
		fmt.Print(err)
		return
	}
	fmt.Print(userInfo)

	var fb FeedBack

	ok, err := checkLogin(userInfo.Username, userInfo.Password)

	if err != nil {
		fmt.Print(err.Error())
		return
	}
	if ok{
		fb.Code = 200
		fb.Ok = true
		fbData,err := json.Marshal(fb)
		fmt.Print(string(fbData))
		if err != nil {
			fmt.Print("反馈数据解析失败")
			return
		}
		fmt.Fprintln(res, string(fbData))
	}else{
		fb.Code = 200
		fb.Ok = false
		fbData,err := json.Marshal(fb)
		fmt.Print(string(fbData))
		if err != nil {
			fmt.Print("反馈数据解析失败")
			return
		}
		fmt.Fprintln(res, string(fbData))
	}

	defer req.Body.Close()

}

func registerHandle(res http.ResponseWriter, req *http.Request){

	res.Header().Set("Access-Control-Allow-Origin", "*") //设置跨域

	result, err := ioutil.ReadAll(req.Body)

	defer req.Body.Close()
	if err!= nil{
		fmt.Print("获取请求参数失败", err.Error())
		return
	}

	var userInfo USER

	err = json.Unmarshal(result, &userInfo)
	if err != nil {
		fmt.Print("解析数据失败", err.Error())
		return
	}

	exist, err := checkExist(userInfo.Username)

	var fb FeedBack
	if exist{
		fmt.Print("用户已存在")
		fb.Code = 200
		fb.Ok = false
	}else{
		ok, err := checkRegister(userInfo.Username, userInfo.Password)

		if err != nil {
			fmt.Print("注册失败", err.Error())
			fb.Code = 200
			fb.Ok = false
		}

		if ok {
			fb.Code = 200
			fb.Ok = true
		}else{
			fb.Code = 200
			fb.Ok = false
		}
	}

	fbData,err := json.Marshal(fb)

	if err != nil{
		fmt.Print("解析反馈信息错误", err.Error())
		return
	}
	fmt.Fprintln(res, string(fbData))

}

func InitDB(){  // 初始化数据库
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, sqlpassword, dbname)

	db,err := sql.Open("postgres", psqlInfo)  // 打开数据库连接

	if err != nil{
		panic(err)
	}

	defer db.Close()

	err = db.Ping()

	if err != nil{
		panic(err)
	}

	fmt.Print("连接成功\n")

	sqlStatement := `CREATE TABLE  IF NOT EXISTS "userinfo"(
		username VARCHAR(25) not null primary key,
		password VARCHAR(25)
	);`

	_, err = db.Exec(sqlStatement)

	if err != nil{
		panic(err)
	}

	fmt.Print("数据库初始化成功!\n")
}

func main() {
	InitDB()
	fmt.Print("开始服务..\n")
	http.HandleFunc("/api/login", loginHandle)
	http.HandleFunc("/api/register", registerHandle)
	http.ListenAndServe(":6661", nil)
	fmt.Print("服务结束\n")
}

