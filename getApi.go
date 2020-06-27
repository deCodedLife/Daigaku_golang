package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

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
		tag, err := getTagID("Р С’Р В»Р С–Р ВµР В±РЎР‚Р В°")
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
		task, err := getHomeTaskID("Р СћР ВµР С•РЎР‚Р С‘РЎРЏ Р С‘Р Р…РЎвЂћР С•РЎР‚Р СР В°РЎвЂ Р С‘Р С‘ Р С—РЎР‚Р В°Р С”РЎвЂљР С‘Р С”Р В° 22")
		log.Println(task)
	*/

	var id int

	err := db.QueryRow("select id from School.HomeTasks where task = ?", task).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func getTestID(task string) (int, error) {

	/*
		Function get tests task id from database using task name
		Return int with id and error
		Example:
		task, err := getTestID("Р СћР ВµР С•РЎР‚Р С‘РЎРЏ Р С‘Р Р…РЎвЂћР С•РЎР‚Р СР В°РЎвЂ Р С‘Р С‘ Р С”Р С•Р Р…РЎвЂљРЎР‚Р С•Р В»РЎРЉР Р…Р В°РЎРЏ 22")
		log.Println(task)
	*/

	var id int

	err := db.QueryRow("select id from School.Tests where task = ?", task).Scan(&id)
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
					json.NewEncoder(w).Encode("Р¤Р°Р№Р» РЅРµ РґРѕСЃС‚СѓРїРµРЅ РґР»СЏ РїСЂРѕСЃРјРѕС‚СЂР°: Doc{id} = " + strconv.Itoa(node.ID))
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
				json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
