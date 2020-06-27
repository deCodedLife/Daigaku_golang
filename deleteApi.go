package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"net/http"
	"os"
)

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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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
					json.NewEncoder(w).Encode("РЈ РІР°СЃ РЅРµС‚ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№. Doc {id} = " + strconv.Itoa(doc.ID))
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
			json.NewEncoder(w).Encode("Р’С‹ РЅРµ РёРјРµРµС‚Рµ СЂР°Р·СЂРµС€РµРЅРёР№ РґР»СЏ РІС‹РїРѕР»РЅРµРЅРёСЏ СЌС‚РёС… РґРµР№СЃС‚РІРёР№")
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