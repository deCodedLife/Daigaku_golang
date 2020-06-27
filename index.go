package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	//"gopkg.in/h2non/bimg.v1"
	"github.com/gorilla/mux"
)

var db *sql.DB             // Database interface
var localimg []os.FileInfo // var for list of files in local dir
// letters for random
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
