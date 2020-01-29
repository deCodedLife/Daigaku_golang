package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type dataBase struct {
	user     string //`json:"user"`
	pass     string //`json:"pass"`
	database string //`json:"database"`
}

type userNode struct {
	id         int    //`json:"id"`
	pass       string //`json:"pass"`
	name       string //`json:"name"`
	status     string //`json:"status"`
	group      string //`json:"group"`
	isCurator  string //`json:"is_curator"`
	profile    string //`json:"profile"`
	curatorTag string //`json:"curatorTag"`
}

type accessForm struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
}

type GroupForm struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	Operator string `json:"operator"`
}

type ImageForm struct {
	Tags  string `json:"tag"`
	Path  string `json:"path"`
	Token string `json:"token"`
}

type MessageForm struct {
	Pick    string `json:"pick"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type TagForm struct {
	Tags    string `json:"tag"`
	Group   string `json:"group"`
	Token   string `json:"token"`
	Curator string `json:"curator"`
}

type TaskForm struct {
	Tags   string `json:"tag"`
	Task   string `json:"task"`
	Group  string `json:"group"`
	Token  string `json:"token"`
	DateTo string `json:"date_to"`
}

type applyTaskForm struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type changeGroupForm struct {
	UserName string `json:"username"`
	Group    string `json:"group"`
	Token    string `json:"token"`
}

type dGroupForm struct {
	Group string `json:"groupname"`
	Token string `json:"token"`
}

type dImage struct {
	Path  string `json:"path"`
	Token string `json:"token"`
}

type dMessage struct {
	Message string `json:"message"`
	Pick    string `json:"pick"`
	Date    string `json:"date"`
	Token   string `json:"token"`
}

type dTag struct {
	Tags  string `json:"tag"`
	Group string `json:"group"`
	Token string `json:"token"`
}

type dTask struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type dUser struct {
	User  string `json:"username"`
	Token string `json:"token"`
}

type gCuratorTagForm struct {
	Tags  string `json:"tag"`
	Token string `json:"token"`
}

type gImages struct {
	Tags  string `json:"tag"`
	Date  string `json:"date"`
	Token string `json:"token"`
}

type imageNode struct {
	id    int
	group string
	tag   string
	date  string
	path  string
}

type gMessages struct {
	Token string `json:"token"`
}

type Message struct {
	id       int
	group    string
	pick     int
	message  string
	operator string
	date     string
}

type gTags struct {
	Group string `json:"group"`
	Token string `json:"token"`
}

type Tag struct {
	id     int
	groups string
	tag    string
	static int
}

type gGroup struct {
	Token string `json:"token"`
}

type Group struct {
	id       int
	name     string
	operator string
	curator  string
}

type gProfile struct {
	Token string `json:"token"`
}

type gUsers struct {
	Types string `json:"type"`
	Group string `json:"group"`
}

type gCurrentUser struct {
	User  string `json:"currentUser"`
	Token string `json:"token"`
}

type gTasks struct {
	Self  string `json:"self"`
	Token string `json:"token"`
}

type Task struct {
	id       int
	group    string
	tag      string
	task     string
	attached string
	dateTo   string
	operator string
	finished int
}

var db *sql.DB
var localimg []os.FileInfo
var localdir = "/var/www/html/school/img"
var database = dataBase{"admin", "8895304025Dr", "School"}
var methods = make(map[string][]string)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createDirectory(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		fmt.Println(dirName, "already exist as a file!")
		return false
	}

	return false
}

func checkToken(token string) (userNode, error) {

	var (
		name string
		pass string
		date string
		err  error

		getTime string
		decoded []byte
		splited []string
	)

	user := userNode{}
	date = time.Now().Format(time.RFC1123)

	splited = strings.SplitAfter(date, " ")
	date = splited[0] + splited[1] + splited[2] + splited[3]

	decoded, err = base64.URLEncoding.DecodeString(token)
	if err != nil {
		return user, errors.New("Can't decode token")
	}

	splited = strings.Split(string(decoded), "|")
	getTime = splited[0]

	pass = splited[1]
	name = splited[2]

	err = db.QueryRow("select `id`, `pass`, `name`, `status`, `fgroup`, `is_curator`, `profile`, `curatorTag` from School.users where name = ?", name).Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.isCurator, &user.profile, &user.curatorTag)
	if err != nil {
		return user, err
	}
	if getTime != date {
		return user, errors.New("Токен устарел")
	}
	if user.pass != pass {
		return user, errors.New("Токен не верен")
	}

	return user, nil

}

func getAccess(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		name string
		pass string
		date string
		summ string
		base string
		user userNode
		form accessForm
		hash = sha256.New()
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	name = form.Name
	pass = form.Pass

	date = time.Now().Format(time.RFC1123)
	splited := strings.SplitAfter(date, " ")
	date = splited[0] + splited[1] + splited[2] + splited[3]
	hash.Write([]byte(pass))
	summ = hex.EncodeToString(hash.Sum(nil))

	result, _ := db.Query("select `id`, `pass`, `name`, `status`, `fgroup`, `is_curator`, `profile`, `curatorTag` from School.users where name = ?", name)
	defer result.Close()
	result.Next()
	_ = result.Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.isCurator, &user.profile, &user.curatorTag)

	if summ != user.pass {
		fmt.Println(user.pass, "|", summ, "|", name)
		fmt.Fprintf(w, "Логин или пароль введен неверно")
		return
	}

	base = base64.URLEncoding.EncodeToString([]byte(date + "|" + summ + "|" + name))
	json.NewEncoder(w).Encode(base)
}

func addGroup(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err      error
		token    string
		group    string
		form     GroupForm
		user     userNode
		operator string // it's name of person
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	token = form.Token
	group = form.Name
	operator = form.Operator

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	if operator != "0" {
		_, err := db.Query("insert into School.groups (`name`,`operator`,'curator') values (?,?,?)", group, user.name, operator)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}
	} else {
		_, err := db.Query("insert into School.groups (`name`,`operator`,'curator') values (?,?,?)", group, user.name, operator)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(w).Encode("succsess")
}

func addImage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		token   string
		image   []byte
		form    ImageForm
		user    userNode
		exist   = false
		err     error
		date    string
		utag    string // User tag. Example: "Информатика"
		path    string // Image path in filesystem
		splited []string
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	utag = form.Tags
	path = form.Path
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if user.status != "updater" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этого")
		return
	}

	image, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	images, err := ioutil.ReadDir(localdir + "/" + user.group + "/")
	createDirectory(localdir + "/" + user.group + "/")
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	hash := sha256.New()
	hash.Write([]byte(image))
	path = hex.EncodeToString(hash.Sum(nil))
	date = time.Now().Format(time.RFC1123)

	splited = strings.SplitAfter(date, " ")
	date = splited[0] + splited[1] + splited[2] + splited[3]

	for _, item := range images {
		if path == item.Name() {
			exist = true
		}
	}

	if exist {
		json.NewEncoder(w).Encode("Похожий файл уже существует")
		return
	}
	ioutil.WriteFile(localdir+"/"+user.group+"/"+path, image, 0777)
	log.Println(user.group)
	_, err = db.Query("insert into School.Images (`groups`,`tag`,`date`,`path`) values (?,?,CURDATE(),?)", user.group, utag, "school/img/"+user.group+"/"+path)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	json.NewEncoder(w).Encode("succsess")

}

func addMessage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err     error
		pick    string
		form    MessageForm
		user    userNode
		message string
		token   string
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	pick = form.Pick
	message = form.Message
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "updater" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("insert into School.messages (`groups`,`pick`,`message`,`operator`,`date`) values (?,?,?,?,CURDATE())", user.group, pick, message, user.name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	json.NewEncoder(w).Encode("succsess")
}

func addTag(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err     error
		tags    string // mean tag. Simple: "Информатика"
		form    TagForm
		user    userNode
		group   string
		token   string
		curator string
	)

	tags = form.Tags
	group = form.Group
	token = form.Token
	curator = form.Curator

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	if curator != "0" {
		_, err := db.Query("update "+database.database+".users set curatorTag = ? where name = ?", tags, curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		_, err = db.Query("insert into School.tags (groups,tag,static) values (?,?,0)", group, tags)
		if err != nil {
			log.Fatalf(err.Error())
			return
		}
	} else {
		_, err := db.Query("insert into School.tags (`groups`,`tag`,`static`) VALUES (?,?,0)", group, tags)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
	}
	json.NewEncoder(w).Encode("succsess")

}

func addTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err    error
		tags   string // I mean one tag
		task   string
		form   TaskForm
		user   userNode
		group  string
		token  string
		dateTo string
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	tags = form.Tags
	task = form.Task
	group = form.Group
	token = form.Token
	dateTo = form.DateTo

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	_, err = db.Query("insert into School.Tasks (`groups`,`tag`,`task`,`attached`,`date_to`,`operator`, `finished`) values (?,?,?,\"false\",?,?,0)", group, tags, task, dateTo, user.name)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

func applyTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		id    int
		err   error
		user  userNode
		form  applyTaskForm
		token string
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	id, _ = strconv.Atoi(form.ID)
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "curator" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("update School.Tasks set `finished` = 1 where `id` = ?", id)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	json.NewEncoder(w).Encode("succsess")

}

func changeGroup(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err      error
		form     changeGroupForm
		user     userNode
		group    string
		token    string
		username string
	)

	json.NewDecoder(r.Body).Decode(&form)

	group = form.Group
	token = form.Token
	username = form.UserName

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("update School.users set fgroup = ? where name = ?", group, username)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	json.NewEncoder(w).Encode("succsess")

}

func deleteGroup(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err   error
		form  dGroupForm
		user  userNode
		group string
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	group = form.Group
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.groups where name = ?", group)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")
}

func deleteImage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err   error
		path  string
		form  dImage
		user  userNode
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	path = form.Path
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "updater" && user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.Images where path = ?", path)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

func deleteMessage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err     error
		pick    int
		form    dMessage
		user    userNode
		date    string
		token   string
		message string
	)

	json.NewDecoder(r.Body).Decode(&form)

	pick, _ = strconv.Atoi(form.Pick)
	date = form.Date
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "updater" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.messages where message = ? and pick = ? and date = ?", message, pick, date)
	if err != nil {
		log.Fatal(err)
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

func deleteTag(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err   error
		tags  string
		form  dTag
		user  userNode
		group string
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	tags = form.Tags
	group = form.Group
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.tags where tag = ? and groups = ?", tags, group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")
}

func deleteTask(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		id    int
		err   error
		form  dTask
		user  userNode
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	id, _ = strconv.Atoi(form.ID)
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "curator" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.Tasks where id = ?", id)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

func deleteUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err   error
		name  string
		form  dUser
		user  userNode
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	name = form.User
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У Вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("delete from School.users where name = ?", name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

// I'm tired. Realy... Now 4:09 Thusday at night, and after few hours i should go to college... I'm fine )
// I'm here. Now 4:00 but now wensday. I can't think normal.

func getCuratorTag(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err   error
		tags  string
		form  gCuratorTagForm
		token string
	)

	json.NewDecoder(r.Body).Decode(&form)

	tags = form.Tags
	token = form.Token

	_, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	userObj := userNode{}
	err = db.QueryRow("select name from School.users where curatorTag = ?", tags).Scan(&userObj.name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode(userObj.name)
}

func getImages(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err      error
		tags     string
		date     string
		form     gImages
		user     userNode
		token    string
		result   *sql.Rows
		imageObj = make(map[string][]map[string]string)
	)

	json.NewDecoder(r.Body).Decode(&form)

	tags = form.Tags
	date = form.Date
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if tags == "0" {
		result, err = db.Query("select * from "+database.database+".Images where groups = ? and date = ?", user.group, date)
	} else {
		result, err = db.Query("select * from "+database.database+".Images where groups = ? and tag = ?", user.group, tags)
	}

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := imageNode{}
		err = result.Scan(&node.id, &node.group, &node.tag, &node.date, &node.path)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["group"] = node.group
		temp["tag"] = node.tag
		temp["date"] = node.date
		temp["image"] = node.path
		imageObj["images"] = append(imageObj["images"], temp)
	}
	print, _ := json.Marshal(imageObj)
	json.NewEncoder(w).Encode(string(print))

}

func getMessages(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err     error
		form    gMessages
		user    userNode
		token   string
		result  *sql.Rows
		message = make(map[string][]map[string]string)
	)

	json.NewDecoder(r.Body).Decode(&form)

	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	result, err = db.Query("select * from School.messages where groups = ?", user.group)
	if err != nil {
		log.Fatal(err.Error())
	}

	for result.Next() {
		node := Message{}
		err = result.Scan(&node.id, &node.group, &node.pick, &node.message, &node.operator, &node.date)
		if err != nil {
			log.Fatal(err.Error())
			return
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
	json.NewEncoder(w).Encode(string(print))

}

func getTags(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "applicaion/json")

	var (
		err    error
		tag    = make(map[string][]map[string]string)
		form   gTags
		user   userNode
		group  string
		token  string
		result *sql.Rows
	)

	json.NewDecoder(r.Body).Decode(&form)

	group = form.Group
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У Вас нет разрешений для этих действий")
		return
	}

	if group != "0" {
		result, err = db.Query("select * from School.tags where groups = ?", group)
	} else {
		result, err = db.Query("select * from School.tags")
	}

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := Tag{}
		err = result.Scan(&node.id, &node.groups, &node.tag, &node.static)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["tag"] = node.tag
		temp["static"] = strconv.Itoa(node.static)
		tag["tags"] = append(tag["tags"], temp)
	}

	print, _ := json.Marshal(tag)
	json.NewEncoder(w).Encode(string(print))

}

func getGroups(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err    error
		form   gGroup
		token  string
		group  = make(map[string][]map[string]string)
		result *sql.Rows
	)

	json.NewDecoder(r.Body).Decode(&form)

	token = form.Token

	_, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	result, err = db.Query("select * from School.groups")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := Group{}
		err = result.Scan(&node.id, &node.name, &node.operator, &node.curator)
		if err != nil {
			log.Fatal(err)
			return
		}

		temp := make(map[string]string)
		temp["group"] = node.name
		temp["operator"] = node.operator
		temp["curator"] = node.curator
		group["groups"] = append(group["groups"], temp)
	}

	print, _ := json.Marshal(group)
	json.NewEncoder(w).Encode(string(print))

}

func getProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err      error
		form     gProfile
		user     userNode
		token    string
		userNode userNode
		userObj  = make(map[string]interface{})
	)

	json.NewDecoder(r.Body).Decode(&form)

	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	err = db.QueryRow("select name, status, fgroup, profile, is_curator, curatorTag  from School.users where name = ?", user.name).Scan(&userNode.name, &userNode.status, &userNode.group, &userNode.profile, &userNode.isCurator, &userNode.curatorTag)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	userObj["name"] = userNode.name
	userObj["status"] = userNode.status
	userObj["group"] = userNode.group
	userObj["profile"] = userNode.profile
	userObj["is_curator"] = userNode.isCurator
	userObj["curatorTag"] = userNode.curatorTag

	print, _ := json.Marshal(userObj)
	json.NewEncoder(w).Encode(string(print))

}

func getUsers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err    error
		form   gUsers
		user   = make(map[string][]string)
		temp   []string
		types  string
		group  string
		groups string
		result *sql.Rows
	)

	json.NewDecoder(r.Body).Decode(&form)

	types = form.Types
	group = form.Group

	if group != "0" {
		groups = " fgroup = \"" + group + "\" and "
	}
	if types == "s" {
		result, err = db.Query("select name from School.users where " + groups + " (status = \"student\" or status = \"updater\")")
	} else if types == "c" {
		result, err = db.Query("select name from School.users where status = \"curator\"")
	} else if types == "a" {
		result, err = db.Query("select name from School.users where status = \"admin\"")
	}

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := userNode{}
		err = result.Scan(&node.name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		temp = append(temp, node.name)
	}

	user["users"] = temp
	print, _ := json.Marshal(user)
	json.NewEncoder(w).Encode(string(print))

}

func getCurrentUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err      error
		form     gCurrentUser
		cuser    string
		token    string
		userNode userNode
		userObj  = make(map[string]string)
	)

	json.NewDecoder(r.Body).Decode(&form)

	cuser = form.User
	token = form.Token

	_, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	err = db.QueryRow("select name, profile from School.users where name = ?", cuser).Scan(&userNode.name, &userNode.profile)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	userObj["username"] = userNode.name
	userObj["profile"] = userNode.profile

	print, _ := json.Marshal(userObj)
	json.NewEncoder(w).Encode(string(print))

}

func getTasks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err    error
		self   string
		form   gTasks
		user   userNode
		task   = make(map[string][]map[string]string)
		token  string
		result *sql.Rows
	)

	json.NewDecoder(r.Body).Decode(&form)

	self = form.Self
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if self != "0" {
		result, err = db.Query("select * from School.Tasks where groups = ? and operator = ?", user.group, user.name)
	} else {
		result, err = db.Query("select * from School.Tasks where groups = ?", user.group)
	}
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := Task{}
		err = result.Scan(&node.id, &node.group, &node.tag, &node.task, &node.attached, &node.dateTo, &node.operator, &node.finished)

		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["id"] = strconv.Itoa(node.id)
		temp["group"] = node.group
		temp["tag"] = node.tag
		temp["task"] = node.task
		temp["attached"] = node.attached
		temp["date_to"] = node.dateTo
		temp["operator"] = node.operator
		temp["finished"] = strconv.Itoa(node.finished)
		task["tags"] = append(task["tags"], temp)
	}

	print, _ := json.Marshal(task)
	json.NewEncoder(w).Encode(string(print))

}

// Hey now 5:47... I'm finished this sh*t

func itemExists(array []string, item string) bool {

	for i := 0; i < len(array); i++ {
		if array[i] == item {
			return true
		}
	}

	return false
}

func init() {

	var err error

	createDirectory(localdir)
	localimg, err = ioutil.ReadDir(localdir)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	db, err = sql.Open("mysql", database.user+":"+database.pass+"@tcp(localhost:3306)/"+database.database)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Connect to database Succsess")
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/access", getAccess).Methods("POST")
	router.HandleFunc("/add-group", addGroup).Methods("POST")
	router.HandleFunc("/add-image", addImage).Methods("POST")
	router.HandleFunc("/add-message", addMessage).Methods("POST")
	router.HandleFunc("/add-tag", addTag).Methods("POST")
	router.HandleFunc("/add-task", addTask).Methods("POST")
	router.HandleFunc("/applyTask", applyTask).Methods("POST")
	router.HandleFunc("/changeGroup", changeGroup).Methods("POST")
	router.HandleFunc("/delete-group", deleteGroup).Methods("POST")
	router.HandleFunc("/delete-image", deleteImage).Methods("POST")
	router.HandleFunc("/delete-message", deleteMessage).Methods("POST")
	router.HandleFunc("/delete-tag", deleteTag).Methods("POST")
	router.HandleFunc("/delete-task", deleteTask).Methods("POST")
	router.HandleFunc("/delete-user", deleteUser).Methods("POST")
	router.HandleFunc("/get-curatorTag", getCuratorTag).Methods("POST")
	router.HandleFunc("/get-images", getImages).Methods("POST")
	router.HandleFunc("/get-messages", getMessages).Methods("POST")
	router.HandleFunc("/get-tags", getTags).Methods("POST")
	router.HandleFunc("/get-groups", getGroups).Methods("POST")
	router.HandleFunc("/get-profile", getProfile).Methods("POST")
	router.HandleFunc("/get-users", getUsers).Methods("POST")
	router.HandleFunc("/get-currentUser", getCurrentUser).Methods("POST")
	router.HandleFunc("/get-tasks", getTasks).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
