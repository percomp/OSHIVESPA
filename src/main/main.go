package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	//security
	"crypto/md5"
)

// app constants
const (
	dbRoot    = "database" + string(filepath.Separator)
	userTable = dbRoot + "users"

	TemplateRoot                = "web" + string(filepath.Separator) + "templates" + string(filepath.Separator)
	StaticURL            string = "" + string(filepath.Separator) + "web" + string(filepath.Separator) + "static" + string(filepath.Separator)
	StaticRoot           string = "web" + string(filepath.Separator) + "static" + string(filepath.Separator)
	CommonRoot                  = ".." + string(filepath.Separator) + ".."
	PracticeRoot                = CommonRoot + string(filepath.Separator) + "OSHIWASP_local" + string(filepath.Separator)
	PracticeInfoFilename        = "oshiwasp_info.xml"

	HidePracticeURL    = "" + string(filepath.Separator) + "hide" + string(filepath.Separator)
	PublishPracticeURL = "" + string(filepath.Separator) + "publish" + string(filepath.Separator)
	RemovePracticeURL  = "" + string(filepath.Separator) + "remove" + string(filepath.Separator)
)

// cookie handling

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

//credentials
type User struct {
	Name string
	Pass [16]byte
}

var userList []User

// Chech credentials
func checkCreds(name string, pass string) (result bool) {
	for _, user := range userList {
		if user.Name == name {
			data := []byte("oshiwasp")
			if user.Pass == md5.Sum(data) {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

type PracticeInfo struct {
	Title          string
	Id             string
	Visibility     bool
	Description    string
	Main_File      string
	AttachmentList []string `xml:"Attachment"`
	LinkList       []string `xml:"Link"`
	Path           string   `xml:"-"`
}

var (
	practiceList []PracticeInfo
)

func setPractices() { //OSHIHORNET
	//restart practice list
	var p []PracticeInfo
	practiceList = p

	//get current practice tree
	d, err := os.Open(PracticeRoot)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()
	files := []string{}
	err = filepath.Walk(PracticeRoot, func(path string, f os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	for _, file := range files {
		if filepath.Base(file) == PracticeInfoFilename {
			xmlFile, err := os.Open(file)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer xmlFile.Close()
			xmlReaded, _ := ioutil.ReadAll(xmlFile)
			var practInfo PracticeInfo
			xml.Unmarshal(xmlReaded, &practInfo)
			practInfo.Path = filepath.Dir(file)
			practiceList = append(practiceList, practInfo)
		}
	}
}

// Context
type Context struct {
	Lang         int
	Static       string
	User         string
	PracticeList []PracticeInfo
}

var context Context

// Start system

func wakeUp() {
	/***TEMP MARSHAL UNMARSHAL EXAMPLE vvv
	var myUser User
	myUser.Name = "oshiwasp"
	data := []byte("oshiwasp")
	myUser.Pass = md5.Sum(data)
	userList = append(userList, myUser)
	log.Println(userList)
	jsondata, _ := json.Marshal(userList)
	log.Println(jsondata)
	var newUserList []User
	json.Unmarshal(jsondata, &newUserList)
	log.Println(newUserList)

	jsonFile, err := os.Create(userTable)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsondata)
	jsonFile.Close()

	//TEMP MARSHAL UNMARSHAL EXAMPLE ^^^**/

	loadUsers()
	setPractices()
	setContext()
}

func loadUsers() {
	content, err := ioutil.ReadFile(userTable)
	if err != nil {
		fmt.Print("Error:", err)
	}
	err = json.Unmarshal(content, &userList)
	if err != nil {
		fmt.Print("Error:", err)
	}
}

func setContext() {
	context.Static = StaticURL
	context.Lang = 1
}

// login page

func loginHandler(response http.ResponseWriter, request *http.Request) {
	t := template.New("login template") // Create a template.
	var loginTemplate string = TemplateRoot + "login.html"
	t, err := template.ParseFiles(loginTemplate)
	if err != nil {
		log.Print("template parsing error: ", err)
	}
	err = t.Execute(response, context)
	if err != nil {
		log.Print("template executing error: ", err)
	}
	//fmt.Fprintf(response, loginPage)
}

// loginSubmit handler

func loginSubmitHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	pass := request.FormValue("password")
	if name != "" && pass != "" {
		if checkCreds(name, pass) {
			setSession(name, response)
		}
	}
	http.Redirect(response, request, "/", 302)
}

// logout handler

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

// render
func render(w http.ResponseWriter, tmpl string, cntxt Context) {
	log.Println("[render]>>>", cntxt)
	//list of templates, put here all the templates needed
	tmplList := []string{fmt.Sprintf("%sbase.html", TemplateRoot),
		//fmt.Sprintf("%smessage.html", TemplateRoot),
		fmt.Sprintf("%s%s.html", TemplateRoot, tmpl)}
	t, err := template.ParseFiles(tmplList...)
	if err != nil {
		log.Print("template parsing error: ", err)
	}
	err = t.Execute(w, cntxt)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

// main page

func indexPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		cntxt := context
		cntxt.User = userName
		cntxt.PracticeList = practiceList
		render(response, "main", cntxt)
	} else {
		http.Redirect(response, request, "/login", 302)
	}
}

// functions
func changeVisibility(practSelected PracticeInfo, visibility bool) {
	log.Println("Change Visibility", practSelected, visibility)
	practSelected.Visibility = visibility
	outputXML, err := xml.MarshalIndent(practSelected, " ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	err2 := os.Remove(practSelected.Path + string(filepath.Separator) + PracticeInfoFilename)
	if err2 != nil {
		fmt.Println(err2)
	}
	xmlFile, err3 := os.Create(practSelected.Path + string(filepath.Separator) + PracticeInfoFilename)
	if err3 != nil {
		fmt.Println(err3)
	}
	defer xmlFile.Close()
	xmlFile.Write(outputXML)
	xmlFile.Close()
}

func deletePractice(practSelected PracticeInfo) {
	//delete Main File
	err1 := os.Remove(practSelected.Path + string(filepath.Separator) + practSelected.Main_File)
	if err1 != nil {
		fmt.Println(err1)
	}
	// delete XML File
	err2 := os.Remove(practSelected.Path + string(filepath.Separator) + PracticeInfoFilename)
	if err2 != nil {
		fmt.Println(err2)
	}
	//Delete Attachments
	for _, attachment := range practSelected.AttachmentList {
		err3 := os.Remove(practSelected.Path + string(filepath.Separator) + attachment)
		if err3 != nil {
			fmt.Println(err3)
		}
	}
	//Delete Directory if empty
	d, err4 := os.Open(practSelected.Path)
	if err4 != nil {
		fmt.Println(err4)
		os.Exit(1)
	}
	defer d.Close()
	files := []string{}
	_ = filepath.Walk(practSelected.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() == false {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		os.RemoveAll(practSelected.Path)
	}
}

func hidePageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		query := strings.Split(request.URL.Path[len(HidePracticeURL):], "/")
		if len(query[0]) != 0 {
			var practSelected PracticeInfo
			for _, practInfo := range practiceList {
				if practInfo.Id == query[0] {
					practSelected = practInfo
					break
				}
			}
			if practSelected.Id != "" {
				log.Println("Hide " + practSelected.Id)
				changeVisibility(practSelected, false)
				setPractices()
				http.Redirect(response, request, "/", http.StatusFound)
			}
			http.NotFound(response, request)
		}
	} else {
		http.Redirect(response, request, "/login", 302)
	}
}

func publishPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		query := strings.Split(request.URL.Path[len(PublishPracticeURL):], "/")
		if len(query[0]) != 0 {
			var practSelected PracticeInfo
			for _, practInfo := range practiceList {
				if practInfo.Id == query[0] {
					practSelected = practInfo
					break
				}
			}
			if practSelected.Id != "" {
				log.Println("Publish " + practSelected.Id)
				changeVisibility(practSelected, true)
				setPractices()
				http.Redirect(response, request, "/", http.StatusFound)
			}
			http.NotFound(response, request)
		}
	} else {
		http.Redirect(response, request, "/login", 302)
	}
}

func removePageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		query := strings.Split(request.URL.Path[len(RemovePracticeURL):], "/")
		if len(query[0]) != 0 {
			var practSelected PracticeInfo
			for _, practInfo := range practiceList {
				if practInfo.Id == query[0] {
					practSelected = practInfo
					break
				}
			}
			if practSelected.Id != "" {
				log.Println("Remove " + practSelected.Id)
				deletePractice(practSelected)
				setPractices()
				http.Redirect(response, request, "/", http.StatusFound)
			}
			http.NotFound(response, request)
		}
	} else {
		http.Redirect(response, request, "/login", 302)
	}
}

//StaticHandler allows to server the statics references
func StaticHandler(w http.ResponseWriter, req *http.Request) {
	staticFile := req.URL.Path[len(StaticURL):]
	if len(staticFile) != 0 {
		f, err := http.Dir(StaticRoot).Open(staticFile)
		if err == nil {
			content := io.ReadSeeker(f)
			http.ServeContent(w, req, staticFile, time.Now(), content)
			return
		}
	}
	http.NotFound(w, req)
}

// server main method

var router = mux.NewRouter()

func main() {

	wakeUp()

	router.HandleFunc("/", indexPageHandler)
	http.HandleFunc(HidePracticeURL, hidePageHandler)
	http.HandleFunc(PublishPracticeURL, publishPageHandler)
	http.HandleFunc(RemovePracticeURL, removePageHandler)

	router.HandleFunc("/login", loginHandler)

	router.HandleFunc("/loginSubmit", loginSubmitHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler)

	http.Handle("/", router)
	http.HandleFunc(StaticURL, StaticHandler)
	http.ListenAndServe(":8689", nil)
}
