package main

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
	Tags  string `json:"tag"`   // Image tag. For example: "Р ВР Р…РЎвЂћР С•РЎР‚Р СР В°РЎвЂљР С‘Р С”Р В°" (varchar)
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
