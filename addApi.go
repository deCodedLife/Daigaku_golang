package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

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
			json.NewEncoder(w).Encode("Р’С‹ РґРѕР»Р¶РЅС‹ СѓРєР°Р·Р°С‚СЊ РіСЂСѓРїРїСѓ")
			return
		}

		user, err = checkToken(token)
		if err != nil {
			fmt.Printf("IP: " + getIP(r) + " try to add user. Token: " + token)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		if user.status != "admin" && user.status != "curator" {
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
			return
		}

		err = db.QueryRow("select name from School.users where name = ?", name).Scan(&temp)
		if temp != "" {
			json.NewEncoder(w).Encode("РџРѕР»СЊР·РѕРІР°С‚РµР»СЊ СѓР¶Рµ СЃСѓС‰РµСЃС‚РІСѓРµС‚")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			utag    string // User tag. Example: "Р ВР Р…РЎвЂћР С•РЎР‚Р СР В°РЎвЂљР С‘Р С”Р В°"
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
			json.NewEncoder(w).Encode("РЈ РІР°СЃ РЅРµС‚ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("РџРѕРґРѕР±РЅС‹Р№ С„Р°Р№Р» СѓР¶Рµ СЃСѓС‰РµСЃС‚РІСѓРµС‚")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("РџСЂРµРґРјРµС‚ СѓР¶Рµ РµСЃС‚СЊ РІ СЌС‚РѕР№ РіСЂСѓРїРїРµ")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р Р°Р·РјРµСЂ С„Р°Р№Р»Р° СЃР»РёС€РєРѕРј РІРµР»РёРє")
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
			json.NewEncoder(w).Encode("РџРѕРґРѕР±РЅС‹Р№ С„Р°Р№Р» СѓР¶Рµ СЃСѓС‰РµСЃС‚РІСѓРµС‚")
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
