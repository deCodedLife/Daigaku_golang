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
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type dataBase struct {
	user     string // Database user
	pass     string // Database password
	database string // Database name
}

// Data database table `users`
type userNode struct {
	id   int    // user id in  (int(13))
	pass string // user password (varchar)
	name string // user name (varchar)
	/*
		Theare types of permissions for user
		admin - can change `groups`, `user` tables
		updater - can change `Images`, `messages`
		curator - can change `Tasks`
		student - can read some tables
	*/
	status     string // Status. Type of permission (varchar)
	group      string // user group name (varchar)
	isCurator  string // Variable for marking user as curator (int)
	profile    string // Profile image path (text)
	curatorTag string // Name of tag if user are curator
}

// some data from database `users`
type accessForm struct {
	Name string `json:"name"` // User name
	Pass string `json:"pass"` // User password
}

type addUserForm struct {
	Pass      string `json:"pass"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Group     string `json:"group"`
	IsCurator string `json:"is_curator"`
	Token     string `json:"token"`
}

// some data from `groups`
type groupForm struct {
	Name     string `json:"name"`     // Group name (varchar)
	Token    string `json:"token"`    // Token
	Operator string `json:"operator"` // Group operator (varchar)
	Curator  string `json:"curator"`  // Group curator
}

// some data from `Images`
type imageForm struct {
	Tags  string `json:"tag"`   // Image tag. For example: "Информатика" (varchar)
	Path  string `json:"path"`  // Abstact image path
	Token string `json:"token"` // Token
}

// some data from `messages`
type messageForm struct {
	Pick    string `json:"pick"`    // pick - It's mark for special messages (int)
	Token   string `json:"token"`   // Token
	Message string `json:"message"` // Message text (text)
}

// some data from `tags`
type tagForm struct {
	Tags    string `json:"tag"`     // Mean one tag. It's name of school subject
	Group   string `json:"group"`   // Name of group which contain this subject
	Token   string `json:"token"`   // Token
	Curator string `json:"curator"` // Curator of this subject
}

// some data from `Tasks`
type taskForm struct {
	Tags   string `json:"tag"`     // Subject which contain task
	Task   string `json:"task"`    // Task
	Group  string `json:"group"`   // group which contain subject which contain task
	Token  string `json:"token"`   // Token
	DateTo string `json:"date_to"` // Date when task will be diskard
}

// some data from `Task` for ApplyTask func
type applyTaskForm struct {
	ID    string `json:"id"`    // ID of task
	Token string `json:"token"` // Token
}

// some data from `groups` and `users`
type changeGroupForm struct {
	UserName string `json:"username"` // User, who will be moved
	Group    string `json:"group"`    // Group name
	Token    string `json:"token"`    // Token
}

// some data from `groups` for deleteGroup func
type dGroupForm struct {
	Group string `json:"groupname"` // Group name
	Token string `json:"token"`     // Token
}

// some data from `Images` for deleteImage func
type dImage struct {
	Path  string `json:"path"`  // Image path
	Token string `json:"token"` // Token
}

// some data from `messages` for deleteMessage func
type dMessage struct {
	Message string `json:"message"` // Message text
	Pick    string `json:"pick"`    // Mark
	Date    string `json:"date"`    // Upload date
	Token   string `json:"token"`   // Token
}

//some data from `tags` for deleteTag func
type dTag struct {
	Tags  string `json:"tag"`   // Tag name. I mean one tag
	Group string `json:"group"` // Name of group
	Token string `json:"token"` // Token
}

//some data from `Tasks` for deleteTask func
type dTask struct {
	ID    string `json:"id"`    // id of task
	Token string `json:"token"` // Token
}

// some data from `users` for deeteUser func
type dUser struct {
	User  string `json:"username"` // User name
	Token string `json:"token"`    // Token
}

// some data from `users` for getCuratorTag func
type gCuratorTagForm struct {
	Tags  string `json:"tag"`   // Tag name. Mean one
	Token string `json:"token"` // Token
}

// some data from `Images` for getImages func
type gImages struct {
	Tags  string `json:"tag"`   // Tag name. Mean one
	Date  string `json:"date"`  // Upload date
	Token string `json:"token"` // Token
}

// some data from `Images` for getImages func
type imageNode struct {
	id      int    // Image id in database
	groupID int    // Group name
	tagID   int    // Tag name
	date    string // upload date name
	path    string // Image path
}

type gMessages struct {
	Token string `json:"token"` // Token
}

// some data from `messages` for getMessages func
type Message struct {
	id         int    // id of message in database
	groupID    int    // Name of group which contain message
	pick       int    // Mark for set special messages
	message    string // Message text
	operatorID int    // Name of user who posted message
	date       string // Upload date
}

// some data from `tags` for getTags func
type gTags struct {
	Group string `json:"group"` // Name of group
	All   string `json:"all"`
	Token string `json:"token"` // Token
}

// some data from `tags` for getTags func
type tag struct {
	id      int    // id of tag in database
	groupID int    // Name of group
	tag     string // Name of tag
	static  int    // Mark for permission to edit it
}

// some data from `groups` for getGroups
type Group struct {
	id         int    // id in database
	name       string // Name of group
	operatorID int    // Name of user who is updater of group
	curatorID  int    // Name of user who is curator of group
}

type gProfile struct {
	User  string `json:"username"` // User name
	Token string `json:"token"`    // Token
}

// some data from `users` for getUser func
type gUsers struct {
	Types string `json:"type"`  // Additional variable
	Group string `json:"group"` // Name of group
}

// some data from `users` for getCurrentUser func
type gCurrentUser struct {
	User  string `json:"currentUser"` // User name
	Token string `json:"token"`       // Token
}

// some data from `Tasks` for getTasks func
type gTasks struct {
	Self  string `json:"self"`  // Additional variable
	Token string `json:"token"` // Token
}

// some data from `Tasks` for getTask func
type Task struct {
	id         int    // id in database
	groupID    int    // Name of group
	tagID      int    // Name of tag
	task       string // Task text
	attached   string // attached (to user *false by defauld)
	dateTo     string // Upload date
	operatorID int    // Name of user who upload it
	finished   int    // Mark for finished state
}

var db *sql.DB                                             // Database interface
var localimg []os.FileInfo                                 // var for list of files in local dir
var localdir = "/var/www/html/school/img"                  // local dir for save images
var database = dataBase{"admin", "8895304025Dr", "School"} // Structure for init database
// letters for random
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func fileExists(filename string) bool {

	/*
		get (filename) in dir
		return bool
		Example:
		FileSystem:
		-- .
		-- ..
		-- dir
		-- abc.jpg
		-- test.png
		state := fileExists("test.png")
		fmt.Println(state)
		Output: true
	*/

	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createDirectory(dirName string) bool {

	/*
		get ( dir name )
		return bool
		Example:
		File System:
		-- .
		-- ..
		state := createDirectory("newDir")
		fmt.Println(state)
		Output: true
		FileSystem:
		-- .
		-- ..
		-- newDir
	*/

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

func getName(id int) (string, error) {

	var name string

	err := db.QueryRow("select name from School.users where id = ?", id).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil

}

func getNameID(name string) (int, error) {

	var id int

	err := db.QueryRow("select id from School.users where name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getTag(id int) (string, error) {

	var tag string

	err := db.QueryRow("select tag from School.tags where id = ?", id).Scan(&tag)
	if err != nil {
		return "", err
	}

	return tag, nil

}

func getTagID(tag string) (int, error) {

	var id int

	err := db.QueryRow("select id from School.tags where tag = ?", tag).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getGroup(id int) (string, error) {

	var group string

	err := db.QueryRow("select name from School.groups where id = ?", id).Scan(&group)
	if err != nil {
		return "", err
	}

	return group, nil

}

func getGroupID(name string) (int, error) {

	var id int

	err := db.QueryRow("select id from School.groups where name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

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

func addUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err     error
		name    string
		pass    string
		form    addUserForm
		user    userNode
		admin   bool
		token   string
		group   string
		status  string
		curator int
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	pass = form.Pass
	if pass == "" {

		for i := 0; i < 17; i++ {
			x := rand.Intn(1)
			if x == 0 {
				pass = pass + string(letters[rand.Intn(len(letters))])
			} else {
				pass = pass + strconv.Itoa(rand.Intn(9))
			}
		}

	}

	group = form.Group
	token = form.Token

	if group == "" {
		json.NewEncoder(w).Encode("Вы должны выбрать группу")
		return
	}

	if form.Token != "" {

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		if user.status != "admin" {
			json.NewEncoder(w).Encode("У вас нет разрешений для этого")
			return
		}

		err = db.QueryRow("select name from School.users where name = ?", name).Scan(&user.name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		if user.name != "" {
			json.NewEncoder(w).Encode("Пользователь с таким именем уже существует")
			return
		}

		status = form.Status
		if status == "" {
			status = "student"
		}

		if form.IsCurator == "" {
			curator = 0
		} else {
			curator, _ = strconv.Atoi(form.IsCurator)
		}

		_, err = db.Query("insert into School.users (`pass`,`name`,`status`,`fgroup`,`is_curator`,`profile`,`curatorTag`) values (?,?,?,?,?,?,\"\",\"\")", pass, name, status, curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

	} else {

		err = db.QueryRow("select name from School.users where name = ?", name).Scan(&user.name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		if user.name != "" {
			json.NewEncoder(w).Encode("Пользователь с таким именем уже существует")
			return
		}

		_, err = db.Query("insert into School.users (`pass`,`name`,`status`,`fgroup`,`is_curator`,`profile`,`curatorTag`) values (?,?,\"student\",?,0,\"\",\"\")", pass, name, group)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

	}

	if !admin {
		json.NewEncoder(w).Encode("succsess")
	} else {
		json.NewEncoder(w).Encode(pass)
	}

}

func addGroup(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err        error
		token      string
		group      string
		form       groupForm
		user       userNode
		curator    string
		operator   string // it's name of person
		curatorID  int
		operatorID int
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	token = form.Token
	group = form.Name
	curator = form.Curator
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

	operatorID, err = getNameID(user.name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	curatorID, err = getNameID(curator)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	_, err = db.Query("insert into School.groups (`name`,`operator`,`curator`) values (?,?,?)", group, curatorID, operator)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")
}

func addImage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		token   string
		image   []byte
		user    userNode
		exist   = false
		param   = mux.Vars(r)
		err     error
		date    string
		utag    string // User tag. Example: "Информатика"
		path    string // Image path in filesystem
		tagID   int
		splited []string
		groupID int
	)

	utagEnc, _ := hex.DecodeString(param["tag"])
	pathEnc, _ := hex.DecodeString(param["path"])

	utag = string(utagEnc)
	path = string(pathEnc)
	token = param["token"]

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
	log.Println(utag)

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

	tagID, err = getTagID(utag)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	groupID, err = getGroupID(user.group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	_, err = db.Query("insert into School.Images (`groupID`,`tagID`,`date`,`path`) values (?,?,CURDATE(),?)", groupID, tagID, "school/img/"+user.group+"/"+path)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	json.NewEncoder(w).Encode("succsess")

}

func addMessage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err        error
		pick       string
		form       messageForm
		user       userNode
		token      string
		message    string
		groupID    int
		operatorID int
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

	groupID, err = getGroupID(user.group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	operatorID, err = getNameID(user.name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	_, err = db.Query("insert into School.messages (`groupID`,`pick`,`message`,`operatorID`,`date`) values (?,?,?,?,CURDATE())", groupID, pick, message, operatorID)
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
		form    tagForm
		user    userNode
		group   string
		token   string
		curator string
		groupID int
	)

	tags = form.Tags
	group = form.Group
	token = form.Token
	curator = form.Curator

	groupID, err = getGroupID(group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
	}

	if user.status != "admin" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	if curator != "0" {
		_, err := db.Query("update School.users set curatorTag = ? where name = ?", tags, curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		_, err = db.Query("insert into School.tags (groups,tag,static) values (?,?,0)", groupID, tags)
		if err != nil {
			log.Fatalf(err.Error())
			return
		}
	} else {
		_, err := db.Query("insert into School.tags (`groups`,`tag`,`static`) VALUES (?,?,0)", groupID, tags)
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
		err        error
		tags       string // I mean one tag
		task       string
		form       taskForm
		user       userNode
		tagID      int
		group      string
		token      string
		dateTo     string
		groupID    int
		operatorID int
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

	groupID, err = getGroupID(group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	tagID, err = getTagID(tags)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	operatorID, err = getNameID(user.name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	_, err = db.Query("insert into School.Tasks (`groupID`,`tagID`,`task`,`attached`,`date_to`,`operatorID`, `finished`) values (?,?,?,\"false\",?,?,0)", groupID, tagID, task, dateTo, operatorID)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode("succsess")

}

func selectTask(w http.ResponseWriter, r *http.Request) {

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

	if user.status != "student" && user.status != "updater" {
		json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
		return
	}

	_, err = db.Query("update School.Tasks set attached = \""+user.name+"|CURDATE()\" where id = ?", id)
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

	_, err = db.Query("update School.Tasks set finished = 1 where id = ?", id)
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
	message = form.Message

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
		err     error
		tags    string
		form    dTag
		user    userNode
		group   string
		token   string
		groupID int
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

	groupID, err = getGroupID(group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	_, err = db.Query("delete from School.tags where tag = ? and groupID = ?", tags, groupID)
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

func getCurator(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/text")

	var (
		err        error
		form       gProfile
		user       userNode
		token      string
		operatorID int
		curatorID  int
		operator   string
		curator    string
	)

	_ = json.NewDecoder(r.Body).Decode(&form)

	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	err = db.QueryRow("select operatorID, curatorId from School.groups where name = ?", user.group).Scan(&operatorID, &curatorID)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	operator, err = getName(operatorID)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	curator, err = getName(curatorID)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	json.NewEncoder(w).Encode("{\"operator\" : \"" + operator + "\", \"curator\" : \"" + curator + "\"}")

}

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
		// special
		group string
		tag   string
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
		result, err = db.Query("select * from School.Images where groups = ? and date = ?", user.group, date)
	} else {
		result, err = db.Query("select * from School.Images where groups = ? and tag = ?", user.group, tags)
	}

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := imageNode{}
		err = result.Scan(&node.id, &node.groupID, &node.tagID, &node.date, &node.path)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		group, err = getGroup(node.groupID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		tag, err = getTag(node.tagID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["group"] = group
		temp["tag"] = tag
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
		// special
		group    string
		operator string
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
		err = result.Scan(&node.id, &node.groupID, &node.pick, &node.message, &node.operatorID, &node.date)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		group, err = getGroup(node.groupID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		operator, err = getName(node.operatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["id"] = strconv.Itoa(node.id)
		temp["group"] = group
		temp["pick"] = strconv.Itoa(node.pick)
		temp["message"] = node.message
		temp["operator"] = operator
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
		all    string
		tags   = make(map[string][]map[string]string)
		form   gTags
		user   userNode
		group  string
		token  string
		result *sql.Rows
	)

	json.NewDecoder(r.Body).Decode(&form)

	all = form.All
	group = form.Group
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if all != "0" {
		if user.status != "admin" {
			json.NewEncoder(w).Encode("У Вас нет разрешений для этих действий")
			return
		}
		result, err = db.Query("select * from School.tags")
	} else {
		if group != "0" {
			result, err = db.Query("select * from School.tags where groups = ?", group)
		} else {
			result, err = db.Query("select * from School.tags")
		}
	}

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := tag{}
		err = result.Scan(&node.id, &node.groupID, &node.tag, &node.static)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["tag"] = node.tag
		temp["static"] = strconv.Itoa(node.static)
		tags["tags"] = append(tags["tags"], temp)
	}

	print, _ := json.Marshal(tags)
	json.NewEncoder(w).Encode(string(print))

}

func getGroups(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var (
		err    error
		group  = make(map[string][]map[string]string)
		result *sql.Rows
		// special
		curator  string
		operator string
	)

	result, err = db.Query("select * from School.groups")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {
		node := Group{}
		err = result.Scan(&node.id, &node.name, &node.operatorID, &node.curatorID)
		if err != nil {
			log.Fatal(err)
			return
		}

		operator, err = getName(node.operatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		curator, err = getName(node.curatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["group"] = node.name
		temp["operator"] = operator
		temp["curator"] = curator
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
		username string
		userObj  = make(map[string]interface{})
	)

	json.NewDecoder(r.Body).Decode(&form)

	token = form.Token
	username = form.User

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if username != "0" {

		if user.status != "admin" {
			json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
			return
		}

		err = db.QueryRow("select name, status, fgroup, profile, is_curator, curatorTag  from School.users where name = ?", username).Scan(&userNode.name, &userNode.status, &userNode.group, &userNode.profile, &userNode.isCurator, &userNode.curatorTag)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

	} else {
		err = db.QueryRow("select name, status, fgroup, profile, is_curator, curatorTag  from School.users where name = ?", user.name).Scan(&userNode.name, &userNode.status, &userNode.group, &userNode.profile, &userNode.isCurator, &userNode.curatorTag)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
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
		// special
		tag      string
		group    string
		groupID  int
		operator string
	)

	json.NewDecoder(r.Body).Decode(&form)

	self = form.Self
	token = form.Token

	user, err = checkToken(token)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	groupID, err = getGroupID(user.group)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	if self != "0" {
		result, err = db.Query("select * from School.Tasks where groupID = ? and operatorID = ?", groupID, user.id)
	} else {
		result, err = db.Query("select * from School.Tasks where groupID = ?", groupID)
	}
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	for result.Next() {

		node := Task{}
		err = result.Scan(&node.id, &node.groupID, &node.tagID, &node.task, &node.attached, &node.dateTo, &node.operatorID, &node.finished)

		if err != nil {
			log.Fatal(err.Error())
			return
		}

		group, err = getGroup(node.groupID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		tag, err = getTag(node.tagID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		operator, err = getName(node.operatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		temp := make(map[string]string)
		temp["id"] = strconv.Itoa(node.id)
		temp["group"] = group
		temp["tag"] = tag
		temp["task"] = node.task
		temp["attached"] = node.attached
		temp["date_to"] = node.dateTo
		temp["operator"] = operator
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
	router.HandleFunc("/add-user", addUser).Methods("POST")
	router.HandleFunc("/add-group", addGroup).Methods("POST")
	router.HandleFunc("/add-image/{tag}/{path}/{token}", addImage).Methods("POST")
	router.HandleFunc("/add-message", addMessage).Methods("POST")
	router.HandleFunc("/add-tag", addTag).Methods("POST")
	router.HandleFunc("/add-task", addTask).Methods("POST")
	router.HandleFunc("/selectTask", selectTask).Methods("POST")
	router.HandleFunc("/applyTask", applyTask).Methods("POST")
	router.HandleFunc("/changeGroup", changeGroup).Methods("POST")
	router.HandleFunc("/delete-group", deleteGroup).Methods("POST")
	router.HandleFunc("/delete-image", deleteImage).Methods("POST")
	router.HandleFunc("/delete-message", deleteMessage).Methods("POST")
	router.HandleFunc("/delete-tag", deleteTag).Methods("POST")
	router.HandleFunc("/delete-task", deleteTask).Methods("POST")
	router.HandleFunc("/delete-user", deleteUser).Methods("POST")
	router.HandleFunc("/get-curator", getCurator).Methods("POST")
	router.HandleFunc("/get-curatorTag", getCuratorTag).Methods("POST")
	router.HandleFunc("/get-images", getImages).Methods("POST")
	router.HandleFunc("/get-messages", getMessages).Methods("POST")
	router.HandleFunc("/get-tags", getTags).Methods("POST")
	router.HandleFunc("/get-groups", getGroups).Methods("POST")
	router.HandleFunc("/get-profile", getProfile).Methods("POST")
	router.HandleFunc("/get-users", getUsers).Methods("POST")
	router.HandleFunc("/get-currentUser", getCurrentUser).Methods("POST")
	router.HandleFunc("/get-tasks", getTasks).Methods("POST")

	for {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
