package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"gopkg.in/h2non/bimg.v1"
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
	Pass   string `json:"pass"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Group  string `json:"group"`
	Token  string `json:"token"`
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
	Tags  string `json:"tag"`   // Image tag. For example: "РРЅС„РѕСЂРјР°С‚РёРєР°" (varchar)
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

// some data for selectUser func
type selectUser struct {
	ID    string `json:"id"`
	User  string `json:"user"`
	Token string `josn:"token"`
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
	Tags    string `json:"tag"`     // Tag name. I mean one tag
	Group   string `json:"group"`   // Name of group
	Token   string `json:"token"`   // Token
	Curator string `json:"curator"` // Teacher
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
type message struct {
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
	curator int    // Who teacher of subject
}

// some data from `groups` for getGroups
type group struct {
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
	Group string `json:"group"` // Group form database
	Token string `json:"token"` // Token
}

// some data from `Tasks` for getTask func
type task struct {
	id         int    // id in database
	groupID    int    // Name of group
	tagID      int    // Name of tag
	task       string // Task text
	dateTo     string // Upload date
	attached   string // who keep it
	operatorID int    // Name of user who upload it
	finished   int    // Mark for finished state
}

type gHomeTask struct {
	ID    string `json:"id"`    // userID
	All   string `json:"all"`   // additional
	Token string `json:"token"` // token
}

type applyHomeTask struct {
	ID    string `json:"id"`    // taskID
	Mark  string `json:"mark"`  // mark
	User  string `json:"user"`  // user
	Type  string `json:"type"`  // method
	Token string `json:"token"` // token
}

type addHomeTask struct {
	ID          string `json:"id"`          // espexial for update function
	Tag         string `json:"tag"`         // tag link
	Docs        string `json:"docs"`        // docs list
	Task        string `json:"task"`        // home task
	Group       string `json:"group"`       // group link
	Token       string `json:"token"`       // token
	DateTo      string `json:"date_to"`     // finish date
	Groups      string `json:"groups"`      // connected groups
	Description string `json:"description"` // comment
}

type homeTask struct {
	id          int    // Task id
	groupID     int    // group id
	tagID       int    // tag id
	task        string // home task
	dateTo      string // finish date
	operatorID  int    // user who create it
	files       string // attached files
	groups      string // attached groups
	description string // description
}

type homeData struct {
	id     int // task id in database
	taskID int // link id to task
	userID int // user id who attached to it
	mark   int // mark.
}

type tests struct {
	id          int    // id in database
	groupID     int    // link to group
	tagID       int    // link to tag
	test        string // Task text
	dateTo      string // finish task date
	operatorID  int    // who create it
	docs        string // docs list
	groups      string // connected groups
	description string // description
}

type testsData struct {
	/* WATNING
	It's simple with home Data
	*/
	// Please use homeData
}

type gDocs struct {
	ID    string `json:"id"`
	Self  string `json:"self"`
	Token string `josn:"token"`
}

type aDocs struct {
	Token      string `json:"token"`
	Comment    string `json:"comment"`
	Permission string `json:"permission"`
}

type changePermission struct {
	ID         string `josn:"id"`
	Permission string `json:"permission"`
	Token      string `json:"token"`
}

type docs struct {
	ID         int    // id in database
	userID     int    // link to user
	ext        string // extention of file
	path       string // path in file system
	date       string // date of upload
	permission int    // permission
	comment    string // comment to file
	name       string // filename
}

var db *sql.DB                                             // Database interface
var localimg []os.FileInfo                                 // var for list of files in local dir
var localdir = "/var/www/html/school/img"                  // local dir for save images
var docfiles = "/var/www/html/school/docs"                 // local dir for save docs
var database = dataBase{"***", "***", "***"} // Structure for init database
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

	/*
		Function get user name from database using user id
		Return string with user name and error
		Example:
		user, err := getName(10)
		log.Println(user)
	*/

	var name string

	err := db.QueryRow("select name from School.users where id = ?", id).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil

}

func getNameID(name string) (int, error) {

	/*
		Function get user id from database using user name
		Return int with id and error
		Example:
		id, err := getNameID("Some name")
		log.Println(id)
	*/

	var id int

	err := db.QueryRow("select id from School.users where name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getTag(id int) (string, error) {

	/*
		Function get name of subject from database using subject id
		Return string with user name and error
		Example:
		tag, err := getTag(10)
		log.Println(tag)
	*/

	var tag string

	err := db.QueryRow("select tag from School.tags where id = ?", id).Scan(&tag)
	if err != nil {
		return "", err
	}

	return tag, nil

}

func getTagID(tag string) (int, error) {

	/*
		Function get subject id from database using subject id
		Return int with id and error
		Example:
		tag, err := getTagID("РђР»РіРµР±СЂР°")
		log.Println(tag)
	*/

	var id int

	err := db.QueryRow("select id from School.tags where tag = ?", tag).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getGroup(id int) (string, error) {

	/*
		Function get group name from database using group id
		Return string with user name and error
		Example:
		group, err := getGroup(10)
		log.Println(group)
	*/

	var group string

	err := db.QueryRow("select name from School.groups where id = ?", id).Scan(&group)
	if err != nil {
		return "", err
	}

	return group, nil

}

func getGroupID(name string) (int, error) {

	/*
		Function get group id from database using group name
		Return int with id and error
		Example:
		group, err := getGroupID("214-f")
		log.Println(group)
	*/

	var id int

	err := db.QueryRow("select id from School.groups where name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getHomeTask(id int) (string, error) {

	/*
		Function get home task from database using task id
		Return string with task and error
		Example:
		task, err := getHomeTask(10)
		log.Println(task)
	*/

	var task string

	err := db.QueryRow("select task from School.HomeTasks where id = ?", id).Scan(&task)
	if err != nil {
		return "", err
	}

	return task, nil

}

func getHomeTaskID(task string) (int, error) {

	/*
		Function get home task id from database using task name
		Return int with id and error
		Example:
		task, err := getHomeTaskID("РўРµРѕСЂРёСЏ РёРЅС„РѕСЂРјР°С†РёРё РїСЂР°РєС‚РёРєР° 22")
		log.Println(task)
	*/

	var id int

	err := db.QueryRow("select id from School.HomeTasks where task = ?", task).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getTest(id int) (string, error) {

	/*
		Function get tests task from database using task id
		Return string with task and error
		Example:
		task, err := getTest(10)
		log.Println(task)
	*/

	var task string

	err := db.QueryRow("select task from School.Tests where id = ?", id).Scan(&task)
	if err != nil {
		return "", err
	}

	return task, nil

}

func getTestID(task string) (int, error) {

	/*
		Function get tests task id from database using task name
		Return int with id and error
		Example:
		task, err := getTestID("РўРµРѕСЂРёСЏ РёРЅС„РѕСЂРјР°С†РёРё РєРѕРЅС‚СЂРѕР»СЊРЅР°СЏ 22")
		log.Println(task)
	*/

	var id int

	err := db.QueryRow("select id from School.Tests where task = ?", task).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

/*
	**********IMPORTANT*********
	Tokens structure and more info

	All passwords which contains in database are hashed with sha256
	Token contain date, user password and user name
	After unification token hashed by base64 (URL)
*/

func checkToken(token string) (userNode, error) {

	/*
		Function validate token and return userNode struct
		In this function token will be decrypted.
		I compare password from database and from token for validate token
		And database response I write to userNode strunc and return back
		Example:
		user, err := checkToken("KWQkQDpRQTeEMTQsn3DJdfBEQO==")
	*/

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
	// Split date for get just date, with out time and etc.
	splited = strings.SplitAfter(date, " ")
	date = splited[0] + splited[1] + splited[2] + splited[3]
	// Decode Token
	decoded, err = base64.URLEncoding.DecodeString(token)
	if err != nil {
		return user, errors.New("Can't decode token")
	}
	// make date unification
	splited = strings.Split(string(decoded), "|")
	if len(splited) == 1 {
		decoded, err = base64.URLEncoding.DecodeString(string(decoded))
		if err != nil {
			return user, errors.New("Can't decode token")
		}
		splited = strings.Split(string(decoded), "|")
	}
	if len(splited) == 1 {
		return user, errors.New("Can't decode token")
	}
	getTime = splited[0]

	pass = splited[1]
	name = splited[2]
	// Get info from database
	err = db.QueryRow("select `id`, `pass`, `name`, `status`, `fgroup`, `is_curator`, `profile`, `curatorTag` from School.users where name = ?", name).Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.isCurator, &user.profile, &user.curatorTag)
	if err != nil {
		return user, err
	}
	if getTime != date {
		return user, errors.New("Токен устарел. Перезагрузите сессию")
	}
	if user.pass != pass {
		return user, errors.New("Аунтификационные данные не верны")
	}

	return user, nil

}

func getIP(r *http.Request) string {
	// Gust return user IP

	forwarded := r.Header.Get("X-FORWARDED-FOR")

	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr

}

func getAccess(w http.ResponseWriter, r *http.Request) {

	/*
		This function generate Token and display it to http response
		Also this function get information from database and compare
		with input data.
		WARNING.
		This function is a part of api and served by mux api
	*/

	defer func() {

		w.Header().Set("Content-Type", "application/text")

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
		// Create date unification
		date = time.Now().Format(time.RFC1123)
		splited := strings.SplitAfter(date, " ")
		date = splited[0] + splited[1] + splited[2] + splited[3]
		hash.Write([]byte(pass))
		summ = hex.EncodeToString(hash.Sum(nil))
		// Get database info
		result, _ := db.Query("select `id`, `pass`, `name`, `status`, `fgroup`, `is_curator`, `profile`, `curatorTag` from School.users where name = ?", name)
		defer result.Close()
		result.Next()
		_ = result.Scan(&user.id, &user.pass, &user.name, &user.status, &user.group, &user.isCurator, &user.profile, &user.curatorTag)
		// Comparing
		if summ != user.pass {
			fmt.Fprintf(w, "Логин или пароль не верны")
			log.Printf("IP: " + getIP(r) + " | user: " + user.name + " try to login")
			return
		}
		// Return token
		base = base64.URLEncoding.EncodeToString([]byte(date + "|" + summ + "|" + name))
		json.NewEncoder(w).Encode([]byte(base))
		log.Printf("IP: " + getIP(r) + " | user: " + user.name + " succsessful logined")

	}()

}

func addUser(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err     error
			name    string
			pass    string
			form    addUserForm
			user    userNode
			ipass   string
			token   string
			group   string
			status  string
			curator int
			// add more
			temp   string
			oper   int
			fgroup int
		)

		_ = json.NewDecoder(r.Body).Decode(&form)

		pass = "fec-college"

		hasher := sha256.New()
		hasher.Write([]byte(pass))
		ipass = hex.EncodeToString(hasher.Sum(nil))

		name = form.Name
		group = form.Group
		token = form.Token

		if group == "" {
			json.NewEncoder(w).Encode("Вы должны указать группу")
			return
		}

		user, err = checkToken(token)
		if err != nil {
			fmt.Printf("IP: " + getIP(r) + " try to add user. Token: " + token)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для этих действий")
			return
		}

		err = db.QueryRow("select name from School.users where name = ?", name).Scan(&temp)
		if temp != "" {
			json.NewEncoder(w).Encode("Пользователь уже существует")
			return
		}

		status = form.Status
		if status == "" {
			status = "student"
			curator = 0
		} else if status == "curator" {
			curator = 1
		} else {
			curator = 0
		}

		_, err = db.Query("insert into School.users (`pass`,`name`,`status`,`fgroup`,`is_curator`,`profile`,`curatorTag`) values (?,?,?,?,?,\"\",\"\")", ipass, name, status, group, curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if status == "updater" {
			oper, err = getNameID(name)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			fgroup, err = getGroupID(group)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			_, err = db.Query("update School.groups set operatorID = ? where id = ?", oper, fgroup)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}
		fmt.Printf("IP: " + getIP(r) + " user " + user.name + " created new user")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func addGroup(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to add group. Token: " + token)
			return
		}

		if user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		operatorID, err = getNameID(operator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		curatorID, err = getNameID(curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		_, err = db.Query("insert into School.groups (`name`,`operator`,`curator`) values (?,?,?)", group, curatorID, operatorID)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		fmt.Printf("IP: " + getIP(r) + " add group: " + group)
		json.NewEncoder(w).Encode("succsess")

	}()

}

func addImage(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			token   string
			image   []byte
			user    userNode
			exist   = false
			param   = mux.Vars(r)
			err     error
			date    string
			utag    string // User tag. Example: "РРЅС„РѕСЂРјР°С‚РёРєР°"
			path    string // Image path in filesystem
			tagID   int
			splited []string
			groupID int
		)

		utagEnc := string(param["tag"])

		utag = string(utagEnc)
		token = param["token"]

		user, err = checkToken(token)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to add image. Token: " + token)
			return
		}

		if user.status != "updater" {
			json.NewEncoder(w).Encode("У вас нет разрешений для этих действий")
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
			json.NewEncoder(w).Encode("Подобный файл уже существует")
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
		fmt.Printf("IP: " + getIP(r) + " upload image")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func addMessage(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to add message. Token: " + token)
			return
		}

		if user.status != "updater" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
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
		fmt.Printf("IP: " + getIP(r) + " create new message")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func addTag(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err     error
			tags    string // mean tag. Simple: "Informatics"
			form    tagForm
			user    userNode
			group   string
			token   string
			curator string
			groupID int
			tagName string
			// special
			curatorID int
			tagID     int
		)

		json.NewDecoder(r.Body).Decode(&form)

		tags = form.Tags
		group = form.Group
		token = form.Token
		curator = form.Curator

		fmt.Println(curator)

		groupID, err = getGroupID(group)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to add tag. Token: " + token)
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			fmt.Printf("IP: " + getIP(r) + " try to add tag. Token: " + token)
			return
		}

		curatorID, err = getNameID(user.name)
		if err != nil {
			log.Fatal(err.Error())
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

		err = db.QueryRow("select name from School.tags where groupID = ? and tag = ? and curatorID = ?", groupID, tagID, curatorID).Scan(tagName)

		if tagName == tags {
			json.NewEncoder(w).Encode("Предмет уже есть в этой группе")
			return
		}

		if curator != "0" {
			var utags string
			err := db.QueryRow("select curatorTag School.users where name = ?", curator).Scan(&utags)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			_, err = db.Query("update School.users set curatorTag = ? where name = ?", utags+" "+tags, curator)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			_, err = db.Query("insert into School.tags (`groupID`,`tag`,`static`,`curatorID`) values (?,?,0,?)", groupID, tags, curatorID)
			if err != nil {
				log.Fatalf(err.Error())
				return
			}
		} else {
			_, err := db.Query("insert into School.tags (`groupID`,`tag`,`static`,`curatorID`) VALUES (?,?,0,?)", groupID, tags, curatorID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}
		fmt.Printf("IP: " + getIP(r) + " create new subject")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func addTask(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to add task. Token: " + token)
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

		fmt.Printf("IP: " + getIP(r) + " create new task")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func selectTask(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to select task. Token: " + token)
			return
		}

		if user.status != "student" && user.status != "updater" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("update School.Tasks set attached = \""+user.name+"|CURDATE()\" where id = ?", id)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return

		}
		fmt.Printf("IP: " + getIP(r) + " select task" + token)
		json.NewEncoder(w).Encode("succsess")

	}()

}

func applyTask(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try apply task. Token: " + token)
			return
		}

		if user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("update School.Tasks set finished = 1 where id = ?", id)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " appled task")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func selectUsers(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id    int
			err   error
			date  string
			form  selectUser
			user  userNode
			cuser string
			token string
			// special
			attached string
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		cuser = form.User
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to select user. Token: " + token)
			return
		}

		if user.status != "curator" && user.status != "updater" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		err = db.QueryRow("select CURDATE()").Scan(&date)
		err = db.QueryRow("select attached from Schoo.Tasks where id = ?", id).Scan(attached)

		if attached != "false" {
			json.NewEncoder(w).Encode("Реферат уже назначен другому пользователю")
			return
		}

		_, err = db.Query("update School.Tasks set attached = ? where id = ?", cuser+"|"+date, id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " select user")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func changeGroup(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id    int
			err   error
			date  string
			form  selectUser
			user  userNode
			cuser string
			token string
			// special
			attached string
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		cuser = form.User
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to select user. Token: " + token)
			return
		}

		if user.status != "curator" && user.status != "updater" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		err = db.QueryRow("select CURDATE()").Scan(&date)
		err = db.QueryRow("select attached from Schoo.Tasks where id = ?", id).Scan(attached)

		if attached != "false" {
			json.NewEncoder(w).Encode("Реферат уже назначен другому пользователю")
			return
		}

		_, err = db.Query("update School.Tasks set attached = ? where id = ?", cuser+"|"+date, id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " select user")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteGroup(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete group. Token: " + token)
			return
		}

		if user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.groups where name = ?", group)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete group")
		json.NewEncoder(w).Encode("succsess")

	}()
}

func deleteImage(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete image. Token: " + token)
			return
		}

		if user.status != "updater" && user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		localPath := strings.Split("/", path)

		err = os.Remove(localdir + "/" + localPath[len(localPath)-1])
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		_, err = db.Query("delete from School.Images where path = ?", path)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete image " + path)
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteMessage(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete message. Token: " + token)
			return
		}

		if user.status != "updater" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.messages where message = ? and pick = ? and date = ?", message, pick, date)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete message")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteTag(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err     error
			tags    string
			form    dTag
			user    userNode
			group   string
			token   string
			groupID int
			curator string
			// special
			curatorID int
		)

		json.NewDecoder(r.Body).Decode(&form)

		tags = form.Tags
		group = form.Group
		token = form.Token
		curator = form.Curator

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to delete tag. Token: " + token)
			return
		}

		curatorID, err = getNameID(curator)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		groupID, err = getGroupID(group)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		_, err = db.Query("delete from School.tags where tag = ? and groupID = ? and curatorID = ?", tags, groupID, curatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete tag " + tags)
		json.NewEncoder(w).Encode("succsess")

	}()
}

func deleteTask(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete task. Token: " + token)
			return
		}

		if user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.Tasks where id = ?", id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete task " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteUser(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete user. Token: " + token)
			return
		}

		if user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.users where name = ?", name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete user " + name)
		json.NewEncoder(w).Encode("succsess")

	}()

}

// I'm tired. Realy... Now 4:09 28.01 Thusday at night, and after few hours I should go to college... I'm fine )
// I'm here. Now 4:00 but now wensday 29.01.20. I can't think normal.
// New line. Today is Wensday/Thirsday 12.02.20. and i should do many things in this code.
// F*ck this code growing day by day. Sh*t...

func getCurator(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err        error
			form       gProfile
			user       userNode
			token      string
			userObj    = make(map[string]string)
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
			fmt.Printf("IP: " + getIP(r) + " try to get curator. Token: " + token)
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

		userObj["operator"] = operator
		userObj["curator"] = curator

		fmt.Printf("IP: " + getIP(r) + " get curator")
		json.NewEncoder(w).Encode(userObj)

	}()

}

func getCuratorTag(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to get curator subject. Token: " + token)
			return
		}

		userObj := userNode{}
		err = db.QueryRow("select name from School.users where curatorTag = ?", tags).Scan(&userObj.name)

		fmt.Printf("IP: " + getIP(r) + " get curator's subject" + token)
		json.NewEncoder(w).Encode(userObj.name)

	}()

}

func getImages(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			tag     string
			tagID   int
			group   string
			groupID int
		)

		json.NewDecoder(r.Body).Decode(&form)

		tags = form.Tags
		date = form.Date
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get images. Token: " + token)
			return
		}

		groupID, err = getGroupID(user.group)

		if tags != "0" {
			tagID, err = getTagID(tags)
			if err != nil {
				json.NewEncoder(w).Encode(err.Error())
				return
			}
			result, err = db.Query("select * from School.Images where groupID = ? and tagID = ?", groupID, tagID)
		} else {
			result, err = db.Query("select * from School.Images where groupID = ? and date = ?", groupID, date)
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

		fmt.Printf("IP: " + getIP(r) + " get iamges")
		json.NewEncoder(w).Encode(imageObj)

	}()

}

func getMessages(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err        error
			form       gMessages
			user       userNode
			token      string
			result     *sql.Rows
			messageObj = make(map[string][]map[string]interface{})
			// special
			group    string
			groupID  int
			operator string
		)

		json.NewDecoder(r.Body).Decode(&form)

		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get messages. Token: " + token)
			return
		}

		groupID, err = getGroupID(user.group)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		result, err = db.Query("select * from School.messages where groupID = ?", groupID)
		if err != nil {
			log.Fatal(err.Error())
		}

		for result.Next() {
			node := message{}
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

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["group"] = group
			temp["pick"] = node.pick
			temp["message"] = node.message
			temp["operator"] = operator
			temp["date"] = node.date
			messageObj["messages"] = append(messageObj["messages"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get messages ")
		json.NewEncoder(w).Encode(messageObj)

	}()

}

func getTags(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "applicaion/json")

		var (
			err    error
			all    string
			tags   = make(map[string][]map[string]interface{})
			form   gTags
			user   userNode
			group  string
			token  string
			result *sql.Rows
			// special
			groupID int
			curator string
		)

		json.NewDecoder(r.Body).Decode(&form)

		all = form.All
		group = form.Group
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get subjects. Token: " + token)
			return
		}

		if all != "0" {
			if user.status != "admin" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			result, err = db.Query("select * from School.tags")
		} else {
			if group != "0" {
				groupID, err = getGroupID(group)
				if err != nil {
					log.Fatal(err.Error())
					return
				}
				result, err = db.Query("select * from School.tags where groupID = ?", groupID)
				if err != nil {
					log.Fatal(err.Error())
					return
				}
			} else {
				result, err = db.Query("select * from School.tags")
				if err != nil {
					log.Fatal(err.Error())
					return
				}
			}
		}

		for result.Next() {
			node := tag{}
			err = result.Scan(&node.id, &node.groupID, &node.tag, &node.static, &node.curator)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			curator, err = getName(node.curator)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			temp := make(map[string]interface{})
			temp["tag"] = node.tag
			temp["curator"] = curator
			temp["static"] = node.static
			tags["tags"] = append(tags["tags"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get subjects")
		json.NewEncoder(w).Encode(tags)

	}()

}

func getGroups(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		type shit struct {
			Curator string `json:"curator"`
		}

		var (
			err      error
			curator  string
			groupObj = make(map[string][]map[string]string)
			result   *sql.Rows
			// special
			ucurator  string
			operator  string
			curatorID int
			// ESPACIAL FUCKING VARS
			thePiceOfShit shit
		)

		json.NewDecoder(r.Body).Decode(&thePiceOfShit)

		ucurator = thePiceOfShit.Curator

		if ucurator != "0" {
			fmt.Println(ucurator)
			curatorID, err = getNameID(ucurator)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			result, err = db.Query("select * from School.groups where curatorID = ?", curatorID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		} else {
			result, err = db.Query("select * from School.groups")
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}

		for result.Next() {
			node := group{}
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
			groupObj["groups"] = append(groupObj["groups"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get groups")
		json.NewEncoder(w).Encode(groupObj)

	}()

}

func getProfile(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err      error
			form     gProfile
			user     userNode
			token    string
			userNode userNode
			username string
			userObj  = make(map[string]string)
		)

		json.NewDecoder(r.Body).Decode(&form)

		token = form.Token
		username = form.User

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get profile. Token: " + token)
			return
		}

		if username != "0" {

			if user.status != "admin" && user.status != "curator" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
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

		fmt.Printf("IP: " + getIP(r) + " get profile")
		json.NewEncoder(w).Encode(userObj)

	}()

}

func getUsers(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
		} else {
			groups = ""
		}
		if types == "s" {
			fmt.Println("select name from School.users where " + groups + " (status = \"student\" or status = \"updater\")")
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
		fmt.Printf("IP: " + getIP(r) + " get users")
		json.NewEncoder(w).Encode(user)

	}()

}

func getCurrentUser(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to get current user. Token: " + token)
			return
		}

		err = db.QueryRow("select name, profile from School.users where name = ?", cuser).Scan(&userNode.name, &userNode.profile)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		userObj["username"] = userNode.name
		userObj["profile"] = userNode.profile

		fmt.Printf("IP: " + getIP(r) + " get current user " + cuser)
		json.NewEncoder(w).Encode(userObj)

	}()

}

func getTasks(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err     error
			self    string
			form    gTasks
			user    userNode
			ugroup  string
			taskObj = make(map[string][]map[string]interface{})
			token   string
			result  *sql.Rows
			// special
			tag      string
			group    string
			groupID  int
			operator string
		)

		json.NewDecoder(r.Body).Decode(&form)

		self = form.Self
		ugroup = form.Group
		token = form.Token

		if ugroup == "" {
			ugroup = "0"
		}

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get tasks. Token: " + token)
			return
		}

		groupID, err = getGroupID(user.group)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if self != "0" {
			if ugroup != "0" {
				groupID, err = getGroupID(ugroup)
				if err != nil {
					log.Fatal(err.Error())
					return
				}
			}
			result, err = db.Query("select * from School.Tasks where groupID = ? and operatorID = ?", groupID, user.id)
		} else {
			if ugroup != "0" {
				fmt.Println(ugroup)
				groupID, err = getGroupID(ugroup)
				if err != nil {
					log.Fatal(err.Error())
					return
				}
			}
			result, err = db.Query("select * from School.Tasks where groupID = ?", groupID)
		}
		/*
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		*/

		for result.Next() {

			node := task{}
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

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["group"] = group
			temp["tag"] = tag
			temp["task"] = node.task
			temp["attached"] = node.attached
			temp["date_to"] = node.dateTo
			temp["operator"] = operator
			temp["finished"] = node.finished
			taskObj["tasks"] = append(taskObj["tasks"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get tasks")
		json.NewEncoder(w).Encode(taskObj)

	}()

}

func uploadIco(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			token string
			image []byte
			user  userNode
			param = mux.Vars(r)
			err   error
		)

		token = param["token"]

		user, err = checkToken(token)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to upload ico. Token: " + token)
			return
		}

		image, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		image, err = bimg.NewImage(image).Resize(256, 256)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		hash := sha256.New()
		hash.Write([]byte(image))

		if user.profile != "" {
			err = os.Remove(localdir + "/" + user.name)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}

		ioutil.WriteFile(localdir+"/"+user.name, image, 0777)

		_, err = db.Query("update School.users set profile = ? where id = ?", "http://95.142.40.58/school/img/"+user.name, user.id)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " upload ico")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func getHomeTasks(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err     error
			self    string
			form    gTasks
			user    userNode
			uGroup  string
			taskObj = make(map[string][]map[string]interface{})
			token   string
			result  *sql.Rows
			// special
			tag      string
			group    string
			groupID  int
			operator string
		)

		json.NewDecoder(r.Body).Decode(&form)

		self = form.Self
		token = form.Token
		uGroup = form.Group

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get home tasks. Token: " + token)
			return
		}

		if uGroup != "0" {
			groupID, err = getGroupID(uGroup)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		} else {
			groupID, err = getGroupID(user.group)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}

		if self != "0" {
			if user.status != "curator" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			result, err = db.Query("select * from School.HomeTasks where groupID = ? and operatorID = ?", groupID, user.id)
		} else {
			result, err = db.Query("select * from School.HomeTasks where groupID = ?", groupID)
		}
		/*
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		*/

		for result.Next() {

			node := tests{}
			err = result.Scan(&node.id, &node.groupID, &node.tagID, &node.test, &node.dateTo, &node.operatorID, &node.docs, &node.groups, &node.description)

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

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["group"] = group
			temp["tag"] = tag
			temp["task"] = node.test
			temp["date_to"] = node.dateTo
			temp["operator"] = operator
			temp["docs"] = node.docs
			temp["description"] = node.description
			taskObj["tests"] = append(taskObj["tests"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get home tasks")
		json.NewEncoder(w).Encode(taskObj)

	}()

}

func deleteHomeTask(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete home task. Token: " + token)
			return
		}

		if user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.HomeTasks where id = ?", id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete home task " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func getHomeData(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			id      int
			err     error
			all     string
			form    gHomeTask
			user    userNode
			taskObj = make(map[string][]map[string]interface{})
			token   string
			result  *sql.Rows
			// special
			tasks string
			guser string
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		all = form.All
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get home data. Token: " + token)
			return
		}

		fmt.Println(id)

		if all != "0" {
			if user.status != "curator" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			result, err = db.Query("select * from School.HomeData where taskID = ? and userID = ?", id, user.id)
		} else {
			result, err = db.Query("select * from School.HomeData where taskID = ?", id)
		}

		for result.Next() {

			node := homeData{}
			err = result.Scan(&node.id, &node.taskID, &node.userID, &node.mark)

			if err != nil {
				log.Fatal(err.Error())
				return
			}

			tasks, err = getHomeTask(node.taskID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			guser, err = getName(node.userID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["task"] = tasks
			temp["user"] = guser
			temp["mark"] = node.mark
			taskObj["tests"] = append(taskObj["tests"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get home data")
		json.NewEncoder(w).Encode(taskObj)

	}()

}

func addHomeTasks(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err        error
			tags       string // I mean one tag
			task       string
			docs       string
			form       addHomeTask
			user       userNode
			tagID      int
			group      string
			token      string
			groups     string
			groupID    int
			operatorID int
		)

		_ = json.NewDecoder(r.Body).Decode(&form)

		tags = form.Tag
		task = form.Task
		docs = form.Docs
		group = form.Group
		token = form.Token
		groups = form.Groups

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to add home task. Token: " + token)
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

		_, err = db.Query("insert into School.HomeTasks (`groupID`,`tagID`,`task`,`date_to`,`operatorID`, `Docs`, `groups`, `description`) values (?,?,?,CURDATE(),?,?,?,?)", groupID, tagID, task, operatorID, docs, groups, form.Description)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + "add home task" + token)
		json.NewEncoder(w).Encode("succsess")

	}()

}

func updateHomeTasks(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("ContentType", "application/text")

		var (
			id          int
			err         error
			form        addHomeTask
			user        userNode
			task        string
			docs        string
			token       string
			description string
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		task = form.Task
		docs = form.Docs
		token = form.Token
		description = form.Description

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error)
			fmt.Printf("IP: " + getIP(r) + " try to update home task. Token: " + token)
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("update School.HomeTasks set task = ?, date_to = CURDATE(), docs = ?, description = ? where id = ?", task, docs, description, id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " update home task " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func applyHome(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id    int
			err   error
			mark  string
			form  applyHomeTask
			user  userNode
			types string
			cuser string
			token string
			// special
			userID int
			task   tests
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		mark = form.Mark
		cuser = form.User
		types = form.Type
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to apply home task. Token: " + token)
			return
		}

		userID, err = getNameID(cuser)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		err = db.QueryRow("select operatorID from School.HomeTasks where id = ?", id).Scan(&task.operatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if user.status != "admin" {
			if user.id != task.operatorID {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
		}

		fmt.Println(mark, id, userID, types)

		if types == "0" {
			_, err = db.Query("insert into School.HomeData (`taskID`,`userID`,`mark`) values (?,?,?)", id, userID, mark)
		} else if types == "1" {
			_, err = db.Query("delete from School.HomeData where taskID = ? and userID = ?", id, userID)
		} else if types == "2" {
			_, err = db.Query("update School.HomeData set mark = ? where taskID = ? and userID = ?", mark, id, userID)
		}
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Printf("IP: " + getIP(r) + " apply home task")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func getTests(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			err     error
			self    string
			form    gTasks
			user    userNode
			uGroup  string
			taskObj = make(map[string][]map[string]interface{})
			token   string
			result  *sql.Rows
			// special
			tag      string
			group    string
			groupID  int
			operator string
		)

		json.NewDecoder(r.Body).Decode(&form)

		self = form.Self
		token = form.Token
		uGroup = form.Group

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get tests. Token: " + token)
			return
		}

		if uGroup != "0" {
			groupID, err = getGroupID(uGroup)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		} else {
			groupID, err = getGroupID(user.group)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}

		if self != "0" {
			if user.status != "curator" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			result, err = db.Query("select * from School.Tests where groupID = ? and operatorID = ?", groupID, user.id)
		} else {
			result, err = db.Query("select * from School.Tests where groupID = ?", groupID)
		}
		/*
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		*/

		for result.Next() {

			node := tests{}
			err = result.Scan(&node.id, &node.groupID, &node.tagID, &node.test, &node.dateTo, &node.operatorID, &node.docs, &node.groups, &node.description)

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

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["group"] = group
			temp["tag"] = tag
			temp["task"] = node.test
			temp["date_to"] = node.dateTo
			temp["operator"] = operator
			temp["docs"] = node.docs
			temp["description"] = node.description
			taskObj["tests"] = append(taskObj["tests"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get tests")
		json.NewEncoder(w).Encode(taskObj)

	}()

}

func getTestsData(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			id      int
			err     error
			all     string
			form    gHomeTask
			user    userNode
			taskObj = make(map[string][]map[string]interface{})
			token   string
			result  *sql.Rows
			// special
			tasks string
			guser string
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		all = form.All
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get tests data. Token: " + token)
			return
		}

		fmt.Println(all, id)

		if all != "0" {
			if user.status != "curator" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			result, err = db.Query("select * from School.TestsData where taskID = ? and userID = ?", id, user.id)
		} else {
			result, err = db.Query("select * from School.TestsData where taskID = ?", id)
		}

		for result.Next() {

			node := homeData{}
			err = result.Scan(&node.id, &node.taskID, &node.userID, &node.mark)

			if err != nil {
				log.Fatal(err.Error())
				return
			}

			fmt.Println(node.id, node.mark, node.taskID, node.userID)

			tasks, err = getTest(node.taskID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			guser, err = getName(node.userID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			temp := make(map[string]interface{})
			temp["id"] = node.id
			temp["task"] = tasks
			temp["user"] = guser
			temp["mark"] = node.mark
			taskObj["tests"] = append(taskObj["tests"], temp)
		}

		fmt.Printf("IP: " + getIP(r) + " get tests data")
		json.NewEncoder(w).Encode(taskObj)

	}()

}

func addTests(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err        error
			tags       string // I mean one tag
			task       string
			docs       string
			form       addHomeTask
			user       userNode
			tagID      int
			group      string
			token      string
			groups     string
			groupID    int
			operatorID int
		)

		_ = json.NewDecoder(r.Body).Decode(&form)

		tags = form.Tag
		task = form.Task
		docs = form.Docs
		group = form.Group
		token = form.Token
		groups = form.Groups

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to add tests. Token: " + token)
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

		_, err = db.Query("insert into School.Tests (`groupID`,`tagID`,`task`,`date_to`,`operatorID`, `Docs`, `groups`, `description`) values (?,?,?,CURDATE(),?,?,?,?)", groupID, tagID, task, operatorID, docs, groups, form.Description)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " add tests")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func updateTests(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("ContentType", "application/text")

		var (
			id          int
			err         error
			form        addHomeTask
			user        userNode
			task        string
			docs        string
			token       string
			description string
		)

		fmt.Println("何だよ")

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		task = form.Task
		docs = form.Docs
		token = form.Token
		description = form.Description

		fmt.Println(task, description, token, docs, id)

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to update tests. Token: " + token)
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("update School.Tests set task = ?, date_to = CURDATE(), Docs = ?, description = ? where id = ?", task, docs, description, id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " update test " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func applyTests(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id    int
			err   error
			mark  string
			form  applyHomeTask
			user  userNode
			types string
			cuser string
			token string
			// special
			userID int
			task   tests
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		mark = form.Mark
		cuser = form.User
		types = form.Type
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to apply tests. Token: " + token)
			return
		}

		if user.status != "curator" && user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		userID, err = getNameID(cuser)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		err = db.QueryRow("select operatorID from School.Tests where id = ?", id).Scan(&task.operatorID)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		if user.status != "admin" {
			if user.id != task.operatorID {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
		}

		fmt.Println(mark, id, userID, types, user)

		if types == "0" {
			_, err = db.Query("insert into School.TestsData (`taskID`,`userID`,`mark`) values (?,?,?)", id, userID, mark)
		} else if types == "1" {
			_, err = db.Query("delete from School.TestsData where taskID = ? and userID = ?", id, userID)
		} else if types == "2" {
			_, err = db.Query("update School.TestsData set mark = ? where taskID = ? and userID = ?", mark, id, userID)
		}
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Printf("IP: " + getIP(r) + " apply test")
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteTests(w http.ResponseWriter, r *http.Request) {

	defer func() {

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
			fmt.Printf("IP: " + getIP(r) + " try to delete test. Token: " + token)
			return
		}

		if user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("delete from School.Tests where id = ?", id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete test " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func getDocs(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/json")

		var (
			id      int
			err     error
			self    string
			docsObj = make(map[string][]map[string]interface{})
			form    gDocs
			user    userNode
			token   string
			result  *sql.Rows
			// special
			cuser  string
			userID int
		)

		json.NewDecoder(r.Body).Decode(&form)

		self = form.Self
		id, _ = strconv.Atoi(form.ID)
		token = form.Token

		user, err = checkToken(token)
		if err != nil {
			log.Fatal(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to get docs. Token: " + token)
			return
		}

		if self != "0" {
			if user.status != "curator" && user.status != "admin" {
				json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
				return
			}
			userID, err = getNameID(user.name)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			result, err = db.Query("select * from School.Docs where userID = ?", userID)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		} else {
			result, err = db.Query("select * from School.Docs where id = ?", id)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		}

		for result.Next() {
			fatal := false
			node := docs{}
			doc := make(map[string]interface{})
			err = result.Scan(&node.ID, &node.ext, &node.path, &node.userID, &node.permission, &node.comment, &node.date, &node.name)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			if node.permission == 0 {
				if user.id != node.userID {
					json.NewEncoder(w).Encode("Файл не доступен для просмотра: Doc{id} = " + strconv.Itoa(node.ID))
					fatal = true
				}
			}
			if !fatal {
				cuser, err = getName(node.userID)
				doc["id"] = node.ID
				doc["user"] = cuser
				doc["ext"] = node.ext
				doc["path"] = node.path
				doc["date"] = node.date
				doc["comment"] = node.comment
				doc["permission"] = node.permission
				doc["name"] = node.name
				docsObj["docs"] = append(docsObj["docs"], doc)
			}

		}

		fmt.Printf("IP: " + getIP(r) + " get docs")
		json.NewEncoder(w).Encode(docsObj)

	}()

}

func addDocs(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			err        error
			ext        string
			name       string
			path       string
			user       userNode
			token      string
			param      = mux.Vars(r)
			comment    string
			permission int
			// special
			userID int
		)

		ext = string(param["ext"])
		name = string(param["name"])
		token = param["token"]
		comment = string(param["comment"])
		permission, _ = strconv.Atoi(string(param["permission"]))

		if ext == " " {
			ext = ""
		}
		if comment == " " {
			comment = ""
		}

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to add doc. Token: " + token)
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		userID, err = getNameID(user.name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		docfile, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		if len(docfile) >= 5368709120 {
			json.NewEncoder(w).Encode("Размер файла слишком велик")
			return
		}

		hash := sha256.New()
		hash.Write(docfile)
		path = string(hex.EncodeToString(hash.Sum(nil)))
		exists := false
		localfiles, _ := ioutil.ReadDir(docfiles)

		for _, item := range localfiles {
			if path == item.Name() {
				exists = true
			}
		}

		if exists {
			json.NewEncoder(w).Encode("Подобный файл уже существует")
			return
		}

		err = ioutil.WriteFile(docfiles+"/"+path, docfile, 0777)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		_, err = db.Query("insert into School.Docs (`ext`,`path`,`userID`,`permission`,`coment`,`date`,`name`) values (?,?,?,?,?,CURDATE(),?)", ext, "http://95.142.40.58/school/docs/"+path, userID, permission, comment, name)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " add doc " + path)
		json.NewEncoder(w).Encode("succsess")

	}()

}

func changePermissions(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id         int
			err        error
			form       changePermission
			user       userNode
			token      string
			permission int
			doc        docs
		)

		json.NewDecoder(r.Body).Decode(&form)

		id, _ = strconv.Atoi(form.ID)
		token = form.Token
		permission, _ = strconv.Atoi(form.Permission)

		user, err = checkToken(token)
		if err != nil {
			json.NewEncoder(w).Encode(err.Error())
			fmt.Printf("IP: " + getIP(r) + " try to change permission. Token: " + token)
			return
		}

		if user.status != "curator" && user.status != "admin" {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		err = db.QueryRow("select name, userID from School.Docs where userID = ?", user.id).Scan(&doc.name, &doc.userID)
		fmt.Println(doc)
		if user.id != doc.userID {
			json.NewEncoder(w).Encode("Вы не имеете разрешений для выполнения этих действий")
			return
		}

		_, err = db.Query("update School.Docs set permission = ? where id = ?", permission, id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " change permission " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

func deleteDoc(w http.ResponseWriter, r *http.Request) {

	defer func() {

		w.Header().Set("Content-Type", "application/text")

		var (
			id    int
			doc   docs
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
			fmt.Printf("IP: " + getIP(r) + " try to delete doc. Token: " + token)
			return
		}

		err = db.QueryRow("select id, name, userID, permission from School.Docs where id = ?", id).Scan(&doc.ID, &doc.name, &doc.userID, &doc.permission)

		if doc.permission == 0 {
			if user.status != "admin" {
				if user.id != doc.userID {
					fmt.Println(doc)
					json.NewEncoder(w).Encode("У вас нет разрешений для этих действий. Doc {id} = " + strconv.Itoa(doc.ID))
					return
				}
			}
		}

		localPath := strings.Split("/", doc.name)

		err = os.Remove(docfiles + "/" + localPath[len(localPath)-1])
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		_, err = db.Query("delete from School.Docs where id = ?", id)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		fmt.Printf("IP: " + getIP(r) + " delete doc " + string(id))
		json.NewEncoder(w).Encode("succsess")

	}()

}

// Hey now 5:47 am... I'm finished this sh*t
// Next day... 7:12 am
// F*ck i just can't go sleep...

func itemExists(array []string, item string) bool {

	for i := 0; i < len(array); i++ {
		if array[i] == item {
			return true
		}
	}

	return false
}

func init() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err error
	createDirectory(localdir)
	localimg, err = ioutil.ReadDir(localdir)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	createDirectory(docfiles)
	createDirectory("logs")

	date := time.Now().Format(time.RFC1123)
	dates := strings.Split(date, ":")
	date = ""
	for i := 0; i < len(dates); i++ {
		date = date + dates[i] + "-"
	}
	f, err := os.OpenFile("logs/"+date+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 777)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	log.Printf("Starting...")

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
	router.HandleFunc("/add-image/{tag}/{token}", addImage).Methods("POST")
	router.HandleFunc("/add-message", addMessage).Methods("POST")
	router.HandleFunc("/add-tag", addTag).Methods("POST")
	router.HandleFunc("/add-task", addTask).Methods("POST")
	router.HandleFunc("/selectTask", selectTask).Methods("POST")
	router.HandleFunc("/applyTask", applyTask).Methods("POST")
	router.HandleFunc("/select-user", selectUsers).Methods("POST")
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
	router.HandleFunc("/upload-ico/{token}", uploadIco).Methods("POST")
	router.HandleFunc("/get-homeTask", getHomeTasks).Methods("POST")
	router.HandleFunc("/get-homeData", getHomeData).Methods("POST")
	router.HandleFunc("/add-homeTask", addHomeTasks).Methods("POST")
	router.HandleFunc("/update-homeTask", updateHomeTasks).Methods("POST")
	router.HandleFunc("/apply-homeData", applyHome).Methods("POST")
	router.HandleFunc("/delete-homeTask", deleteHomeTask).Methods("POST")
	router.HandleFunc("/get-tests", getTests).Methods("POST")
	router.HandleFunc("/get-testsData", getTestsData).Methods("POST")
	router.HandleFunc("/add-tests", addTests).Methods("POST")
	router.HandleFunc("/update-test", updateTests).Methods("POST")
	router.HandleFunc("/apply-testsData", applyTests).Methods("POST")
	router.HandleFunc("/delete-tests", deleteTests).Methods("POST")
	router.HandleFunc("/get-docs", getDocs).Methods("POST")
	router.HandleFunc("/add-docs/{name}/{comment}/{permission}/{ext}/{token}", addDocs).Methods("POST")
	router.HandleFunc("/change-permission", changePermissions).Methods("POST")
	router.HandleFunc("/delete-docs", deleteDoc).Methods("POST")

	defer func() {
		log.Fatal(http.ListenAndServe(":8080", router))
	}()
}
