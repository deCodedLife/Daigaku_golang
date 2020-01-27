package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DataBase struct {
	user     string
	pass     string
	database string
}

type User struct {
	id         int
	pass       string
	name       string
	status     string
	group      string
	is_curator string
	profile    string
	curatorTag string
}

type Structure struct {
	name  string
	value string
}

var db *sql.DB
var methods = make(map[string][]string)
var Database = DataBase{"admin", "8895304025Dr", "School"}

func getArgs(api string) ([]string, error) {

	if len(methods[api]) == 0 {
		return []string(methods["null"]), errors.New("Неизвестный метод")
	}

	return methods[api], nil
}

func getAccess(name string, pass string, time string, Writer http.ResponseWriter) error {
	user := User{}
	hasher := sha256.New()
	err := db.QueryRow("select * from "+Database.database+".users where `name` = ?", name).Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.is_curator, &user.profile, &user.curatorTag)
	if err != nil {
		return nil
	}
	hasher.Write([]byte(pass))
	if hex.EncodeToString(hasher.Sum(nil)) != user.pass {
		fmt.Fprintf(Writer, "Логин или пароль введен неверно")
		return errors.New("Логин или пароль введен неверно")
	}
	baseString := ""
	baseString = baseString + time + "|" + hex.EncodeToString(hasher.Sum(nil)) + "|" + name
	fmt.Fprintf(Writer, "%s", base64.URLEncoding.EncodeToString([]byte(baseString)))
	return nil
}
func addGroup(groupname string, operator string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	if operator != "0" {
		_, err := db.Query("insert into "+Database.database+".groups (`name`,`operator`,'curator') values (?,?,?)", groupname, user.name, operator)
		if err != nil {
			return err
		}
	} else {
		_, err := db.Query("insert into "+Database.database+".groups (`name`,`operator`,'curator') values (?,?,?)", groupname, user.name, operator)
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func addImage(tag string, path string, image []byte, user User, Writer http.ResponseWriter) error {
	//data := []Structure{}
	//node0 := Structure{"tag", tag}
	//node1 := Structure{"path", path}
	//return nil
	return nil
}
func addMessage(pick string, message string, date string, user User, Writer http.ResponseWriter) error {
	if user.status != "updater" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("insert into "+Database.database+".Messages (`group`,`pick`,`message`,`operator`,`date`) values (?,?,?,?,?)", user.group, pick, message, user.name, date)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func addTag(tag string, curator string, group string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	if curator != "0" {
		_, err := db.Query("update "+Database.database+".users set `curatorTag` = ? where name = ?", tag, curator)
		if err != nil {
			return err
		}
		_, err = db.Query("insert into "+Database.database+".tags (`groups`,`tag`,`static`) values (?,?,?)", group, tag, 0)
		if err != nil {
			return err
		}
	} else {
		_, err := db.Query("insert into "+Database.database+".tags (`groups`,`tag`,`static`) values (?,?,?)", user.group, tag, 0)
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func applyTask(id string, user User, Writer http.ResponseWriter) error {
	if user.status != "curator" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("update "+Database.database+".Tasks set `finished` = 1 where `id` = ?", id)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func changeGroup(username string, group string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("update "+Database.database+".users set fgroup = ? where `name` = ?", group, username)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteGroup(groupname string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("delete * from "+Database.database+".groups where `name` = ?", groupname)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteImage(path string, user User, Writer http.ResponseWriter) error {
	if user.status != "updater" && user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("delete * from "+Database.database+".Images where `path` = ?", path)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteMessage(message string, pick string, date string, user User, Writer http.ResponseWriter) error {
	if user.status != "updater" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("delete * from "+Database.database+".messages where `message` = ? and pick` = ? and `date = ?", message, pick, message)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteTag(tag string, group string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("delete * from "+Database.database+".tags where `tag` = ? and `group` = ?", tag, group)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteTask(id int, user User, Writer http.ResponseWriter) error {
	if user.status != "curator" {
		return errors.New("У вас нет разрешений для этих действий")
	}
	_, err := db.Query("delete * from "+Database.database+".Tasks where `id` = ?", id)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
}
func deleteUser(username string, user User, Writer http.ResponseWriter) error {
	if user.status != "admin" {
		return errors.New("У Вас нет ")
	}
	_, err := db.Query("delete * from "+Database.database+".users where `name` = ?", username)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, "succsess")
	return nil
} // I'm tired. Realy... Now 4:09 at night, and after few hours i should go to college... I'm fine )
func getCuratorTag(tag string, user User, Writer http.ResponseWriter) error {
	user_ := User{}
	err := db.QueryRow("select * from "+Database.database+".users"+" where curatorTag = ?", tag).Scan(&user_.id, &user_.pass, &user_.name, &user_.status, &user_.group, &user_.is_curator, &user_.profile, &user_.curatorTag)
	if err != nil {
		return err
	}
	fmt.Fprintf(Writer, user_.name)
	return nil
}
func getImages(tag string, date string, user User, Writer http.ResponseWriter) error {
	type Image_ struct {
		id    int
		group string
		tag   string
		date  string
		path  string
	}
	var result *sql.Rows
	var err error
	image_ := make(map[string][]map[string]string)
	image_["images"] = []map[string]string{}
	if tag == "0" {
		result, err = db.Query("select * from "+Database.database+".Images where groups = ? and date = ?", user.group, date)
	} else {
		result, err = db.Query("select * from "+Database.database+".Images where groups = ? and tag = ?", user.group, tag)
	}
	if err != nil {
		return err
	}
	for result.Next() {
		node := Image_{}
		err = result.Scan(&node.id, &node.group, &node.tag, &node.date, &node.path)
		if err != nil {
			return err
		}
		temp := make(map[string]string)
		temp["group"] = node.group
		temp["tag"] = node.tag
		temp["date"] = node.date
		temp["image"] = node.path
		image_["images"] = append(image_["images"], temp)
	}
	print, _ := json.Marshal(image_)
	fmt.Println(string(print))
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getMessages(user User, Writer http.ResponseWriter) error {
	type Message struct {
		id       int
		group    string
		pick     int
		message  string
		operator string
		date     string
	}
	message := make(map[string][]map[string]string)
	result, err := db.Query("select * from "+Database.database+".messages where groups = ?", user.group)
	if err != nil {
		return err
	}

	for result.Next() {
		node := Message{}
		err = result.Scan(&node.id, &node.group, &node.pick, &node.message, &node.operator, &node.date)
		if err != nil {
			return err
		}
		temp := make(map[string]string)
		temp["id"] = strconv.Itoa(node.id)
		temp["group"] = node.group
		temp["pick"] = strconv.Itoa(node.pick)
		temp["message"] = node.message
		temp["operator"] = node.operator
		temp["date"] = node.date
		message["messages"] = append(message["messages"], temp)
	}
	print, _ := json.Marshal(message)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getTags(group string, user User, Writer http.ResponseWriter) error {
	type Tag struct {
		id     int
		groups string
		tag    string
		static int
	}
	var result *sql.Rows
	var err error
	tag := make(map[string][]map[string]string)
	if user.status != "admin" {
		return errors.New("У Вас нет разрешений для этих действий")
	}
	if group != "0" {
		result, err = db.Query("select * from "+Database.database+".tags where groups = ?", group)
	} else {
		result, err = db.Query("select * from " + Database.database + ".tags")
	}

	if err != nil {
		return err
	}
	for result.Next() {
		node := Tag{}
		err = result.Scan(&node.id, &node.groups, &node.tag, &node.static)
		if err != nil {
			return err
		}
		temp := make(map[string]string)
		temp["tag"] = node.tag
		temp["static"] = strconv.Itoa(node.static)
		tag["tags"] = append(tag["tags"], temp)
	}
	print, _ := json.Marshal(tag)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getGroups(Writer http.ResponseWriter) error {
	type Group struct {
		id       int
		name     string
		operator string
		curator  string
	}
	group := make(map[string][]map[string]string)
	result, err := db.Query("select * from " + Database.database + ".groups")
	if err != nil {
		return err
	}

	for result.Next() {
		node := Group{}
		err = result.Scan(&node.id, &node.name, &node.operator, &node.curator)
		if err != nil {
			return err
		}
		temp := make(map[string]string)
		temp["group"] = node.name
		temp["operator"] = node.operator
		temp["curator"] = node.curator
		group["groups"] = append(group["groups"], temp)
	}
	print, _ := json.Marshal(group)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getProfile(user User, Writer http.ResponseWriter) error {
	_user := User{}
	user_ := make(map[string]interface{})
	err := db.QueryRow("select * from "+Database.database+".users where name = ?", user.name).Scan(&_user.id, &_user.pass, &_user.name, &_user.status, &_user.group, &_user.is_curator, &_user.profile, &_user.curatorTag)
	if err != nil {
		return err
	}
	user_["name"] = _user.name
	user_["status"] = _user.status
	user_["group"] = _user.group
	user_["profile"] = _user.profile
	user_["is_curator"] = _user.is_curator
	user_["curatorTag"] = _user.curatorTag
	print, _ := json.Marshal(user_)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getUsers(types string, group string, Writer http.ResponseWriter) error {
	_user := make(map[string][]string)
	_temp := []string{}
	var result *sql.Rows
	var err error
	var groups string = ""
	if group != "" {
		groups = " fgroup = " + group + " and "
	}
	if types == "s" {
		result, err = db.Query("select * from " + Database.database + ".users where " + groups + " (status = \"student\" or status = \"updater\")")
	} else if types == "c" {
		result, err = db.Query("select * from " + Database.database + ".users where status = \"curator\"")
	} else if types == "a" {
		result, err = db.Query("select * from " + Database.database + ".users where status = \"admin\"")
	}
	if err != nil {
		return err
	}
	for result.Next() {
		node := User{}
		err = result.Scan(&node.id, &node.pass, &node.name, &node.status, &node.group, &node.is_curator, &node.profile, &node.curatorTag)
		if err != nil {
			return err
		}
		_temp = append(_temp, node.name)
	}
	_user["users"] = _temp
	print, _ := json.Marshal(_user)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getCurrentUser(currentUser string, user User, Writer http.ResponseWriter) error {
	_user := User{}
	user_ := make(map[string]string)
	err := db.QueryRow("select * from "+Database.database+".users where name = ?", currentUser).Scan(&_user.id, &_user.pass, &_user.name, &_user.status, &_user.group, &_user.is_curator, &_user.profile, &_user.curatorTag)
	if err != nil {
		return err
	}
	user_["username"] = _user.name
	user_["profile"] = _user.profile
	print, _ := json.Marshal(user_)
	fmt.Fprintf(Writer, string(print))
	return nil
}
func getTasks(self string, user User, Writer http.ResponseWriter) error {
	type Task struct {
		id       int
		group    string
		tag      string
		task     string
		attached string
		date_to  string
		operator string
		finished int
	}
	task := make(map[string][]map[string]string)
	var result *sql.Rows
	var err error
	if self != "0" {
		result, err = db.Query("select * from "+Database.database+".Tasks where groups = ? and operator = ?", user.group, user.name)
	} else {
		result, err = db.Query("select * from "+Database.database+".Tasks where groups = ?", user.group)
	}
	if err != nil {
		return err
	}
	for result.Next() {
		node := Task{}
		err = result.Scan(&node.id, &node.group, &node.tag, &node.task, &node.attached, &node.date_to, &node.operator, &node.finished)
		if err != nil {
			return err
		}
		temp := make(map[string]string)
		temp["id"] = strconv.Itoa(node.id)
		temp["group"] = node.group
		temp["tag"] = node.tag
		temp["task"] = node.task
		temp["attached"] = node.attached
		temp["date_to"] = node.date_to
		temp["operator"] = node.operator
		temp["finished"] = strconv.Itoa(node.finished)
		task["tags"] = append(task["tags"], temp)
	}
	print, _ := json.Marshal(task)
	fmt.Fprintf(Writer, string(print))
	return nil
}

func itemExists(array []string, item string) bool {

	for i := 0; i < len(array); i++ {
		if array[i] == item {
			return true
		}
	}

	return false
}

func HttpServer(Writer http.ResponseWriter, Request *http.Request) {

	baseURL := Request.URL.Path[1:]
	splited := strings.Split(baseURL, "/")

	api := splited[0]
	var allParams []string
	var data = time.Now().Format(time.RFC1123)

	dataSplited := strings.SplitAfter(data, " ")
	data = dataSplited[0] + dataSplited[1] + dataSplited[2] + dataSplited[3]

	params, err := getArgs(api)

	if err != nil {
		fmt.Fprintf(Writer, err.Error())
		return
	}

	if len(splited) < len(params) {
		fmt.Fprintf(Writer, "Вы должны указать все параметры")
		return
	}

	if len(splited) > (len(params) + 2) {
		fmt.Fprintf(Writer, "Слишком много параметров")
		return
	}

	for i := 0; i < len(params); i++ {
		allParams = append(allParams, splited[i+1])
	}

	db, err = sql.Open("mysql", Database.user+":"+Database.pass+"@tcp(localhost:3306)/"+Database.database)
	if err != nil {
		fmt.Fprintf(Writer, err.Error())
		fmt.Println(err.Error())
		return
	} else {
		fmt.Println("Connect to database Succsess")
	}

	if itemExists(params, "token") {
		prepare := allParams[len(allParams)-1]
		decoded, _ := base64.URLEncoding.DecodeString(prepare)
		splitDecoded := strings.Split(string(decoded), "|")
		time := splitDecoded[0]
		pass := splitDecoded[1]
		name := splitDecoded[2]
		if time != data {
			fmt.Fprintf(Writer, "Токен устарел")
			return
		}
		user := User{}
		err := db.QueryRow("select * from "+Database.database+".users where `name` = ?", name).Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.is_curator, &user.profile, &user.curatorTag)
		if err != nil {
			fmt.Fprintf(Writer, err.Error())
			return
		}
		if user.pass != pass {
			fmt.Fprintf(Writer, "Токен не верен")
			return
		}

		switch api {
		case "add-group":
			err = addGroup(allParams[0], allParams[1], user, Writer)
			fmt.Println(err)
		case "add-image":
			err = addImage(allParams[0], allParams[1], []byte(allParams[2]), user, Writer)
			fmt.Println(err)
		case "add-message":
			err = addMessage(allParams[0], allParams[1], allParams[2], user, Writer)
			fmt.Println(err)
		case "add-tag":
			err = addTag(allParams[0], allParams[1], allParams[2], user, Writer)
			fmt.Println(err)
		case "apply-Task":
			err = applyTask(allParams[0], user, Writer)
			fmt.Println(err)
		case "change-group":
			err = changeGroup(allParams[0], allParams[1], user, Writer)
			fmt.Println(err)
		case "delete-group":
			err = deleteGroup(allParams[0], user, Writer)
			fmt.Println(err)
		case "delete-image":
			err = deleteImage(allParams[0], user, Writer)
			fmt.Println(err)
		case "delete-messages":
			err = deleteMessage(allParams[0], allParams[1], allParams[2], user, Writer)
			fmt.Println(err)
		case "get-tags":
			err = getTags(allParams[0], user, Writer)
			fmt.Println(err)
		case "get-profile":
			err = getProfile(user, Writer)
			fmt.Println(err)
		case "get-currentUser":
			err = getCurrentUser(allParams[0], user, Writer)
			fmt.Println(err)
		case "get-tasks":
			err = getTasks(allParams[0], user, Writer)
			fmt.Println(err)
		case "get-messages":
			err = getMessages(user, Writer)
			fmt.Println(err)
		case "get-images":
			err = getImages(allParams[0], allParams[1], user, Writer)
			fmt.Println(err)
		}
	} else {
		switch api {
		case "access":
			getAccess(allParams[0], allParams[1], data, Writer)
		case "get-groups":
			getGroups(Writer)
		case "get-users":
			getUsers(allParams[0], allParams[1], Writer)
		}
	}
	//

	defer db.Close()
}

func main() {

	methods["access"] = []string{"name", "pass"}
	methods["add-group"] = []string{"name", "operator", "token"}
	methods["add-image"] = []string{"tag", "path", "image", "token"}
	methods["add-message"] = []string{"pick", "message", "date", "token"}
	methods["add-tag"] = []string{"tag", "curator", "group", "token"}
	methods["add-task"] = []string{"tag", "group", "task", "date_to", "token"}
	methods["apply-Task"] = []string{"id", "token"}
	methods["change-group"] = []string{"username", "group", "token"}
	methods["delete-group"] = []string{"groupname", "token"}
	methods["delete-image"] = []string{"path", "token"}
	methods["delete-message"] = []string{"message", "pick", "date"}
	methods["delete_tag"] = []string{"tag", "group", "token"}
	methods["delete-task"] = []string{"id", "token"}
	methods["delete-user"] = []string{"username", "token"}
	methods["get-curatorTag"] = []string{"tag", "token"}
	methods["get-images"] = []string{"tag", "date", "token"}
	methods["get-messages"] = []string{"token"}
	methods["get-tags"] = []string{"group", "token"}
	methods["get-groups"] = []string{"some"}
	methods["get-profile"] = []string{"token"}
	methods["get-users"] = []string{"type", "group"}
	methods["get-currentUser"] = []string{"currentUser", "token"}
	methods["get-tasks"] = []string{"self", "token"}
	methods["null"] = []string{""}

	http.HandleFunc("/", HttpServer)
	http.ListenAndServe(":8080", nil)

}
