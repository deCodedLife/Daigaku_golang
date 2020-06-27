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

	"github.com/gorilla/mux"
)

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
		return user, errors.New("РўРѕРєРµРЅ СѓСЃС‚Р°СЂРµР». РџРµСЂРµР·Р°РіСЂСѓР·РёС‚Рµ СЃРµСЃСЃРёСЋ")
	}
	if user.pass != pass {
		return user, errors.New("РђСѓРЅС‚РёС„РёРєР°С†РёРѕРЅРЅС‹Рµ РґР°РЅРЅС‹Рµ РЅРµ РІРµСЂРЅС‹")
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
			fmt.Fprintf(w, "Р›РѕРіРёРЅ РёР»Рё РїР°СЂРѕР»СЊ РЅРµ РІРµСЂРЅС‹")
			log.Printf("IP: " + getIP(r) + " | user: " + user.name + " try to login")
			return
		}
		// Return token
		base = base64.URLEncoding.EncodeToString([]byte(date + "|" + summ + "|" + name))
		json.NewEncoder(w).Encode([]byte(base))
		log.Printf("IP: " + getIP(r) + " | user: " + user.name + " succsessful logined")

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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
			return
		}

		err = db.QueryRow("select CURDATE()").Scan(&date)
		err = db.QueryRow("select attached from Schoo.Tasks where id = ?", id).Scan(attached)

		if attached != "false" {
			json.NewEncoder(w).Encode("Р РµС„РµСЂР°С‚ СѓР¶Рµ РЅР°Р·РЅР°С‡РµРЅ РґСЂСѓРіРѕРјСѓ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
			return
		}

		err = db.QueryRow("select CURDATE()").Scan(&date)
		err = db.QueryRow("select attached from Schoo.Tasks where id = ?", id).Scan(attached)

		if attached != "false" {
			json.NewEncoder(w).Encode("Р РµС„РµСЂР°С‚ СѓР¶Рµ РЅР°Р·РЅР°С‡РµРЅ РґСЂСѓРіРѕРјСѓ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ")
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

// I'm tired. Realy... Now 4:09 28.01 Thusday at night, and after few hours I should go to college... I'm fine )
// I'm here. Now 4:00 but now wensday 29.01.20. I can't think normal.
// New line. Today is Wensday/Thirsday 12.02.20. and i should do many things in this code.
// F*ck this code growing day by day. Sh*t...

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
		/*
			image, err = bimg.NewImage(image).Resize(256, 256)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
		*/
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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

		fmt.Println("дЅ•гЃ г‚€")

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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
			return
		}

		err = db.QueryRow("select name, userID from School.Docs where userID = ?", user.id).Scan(&doc.name, &doc.userID)
		fmt.Println(doc)
		if user.id != doc.userID {
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
