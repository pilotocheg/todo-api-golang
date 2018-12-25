package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/coolbed/mgo-oid"
	"github.com/subosito/gotenv"
	"github.com/user/todo-golang/db"
	hp "github.com/user/todo-golang/helpers"
)

// PORT for connection
var PORT string

//create new struct witg mongoDb connection and methods for it
var svc = new(db.Svc)

// ALIVE is timestamp for server's alive check
var ALIVE = time.Now().Format("2006-01-02 15:04:05")

//Envs is a map of Envs variables
var Envs = make(map[string]string)

//this func gets all envs values to a "Envs" map
func prepareEnvs() {
	gotenv.Load()
	Envs["PORT"] = hp.EnvCheck("PORT")
	Envs["DYNAMODB_ENDPOINT"] = hp.EnvCheck("DYNAMODB_ENDPOINT")
	Envs["AWS_REGION"] = hp.EnvCheck("AWS_REGION")
	Envs["TABLE_NAME"] = hp.EnvCheck("TABLE_NAME")
}

//this func sets up flag values to local variables
func processFlags() {
	flag.StringVar(&PORT, "port", Envs["PORT"], "Set port value")
	flag.Parse()
}

// It is struct for json response at "/" path
type upTimeCheck struct {
	Alive     bool   `json:"alive"`
	AliveAt   string `json:"aliveAt"`
	Timestamp string `json:"timestamp"`
}

//creating main http routes
type route struct {
	pattern *regexp.Regexp
	method  string
	handler http.Handler
}

type regexpHandler struct {
	routes []*route
}

func (h *regexpHandler) HandleFunc(r string, m string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, m, http.HandlerFunc(handler)})
}

func (h *regexpHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(req.URL.Path) && route.method == req.Method {
			route.handler.ServeHTTP(res, req)
			return
		}
	}
	http.NotFound(res, req)
}

// CONTROLLERS FOR ROUTES
func homeController(res http.ResponseWriter, req *http.Request) {
	um := &upTimeCheck{
		Alive:     true,
		AliveAt:   ALIVE,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	res.WriteHeader(200)
	hp.JSONResponse(res, *um)
	req.Body.Close()
}

func createNew(res http.ResponseWriter, req *http.Request) {
	item := db.TodoItem{}

	if err := hp.JSONDecode(req.Body, &item); hp.ErrorCheck(res, err, 500) {
		return
	}
	item.ID = oid.NewOID().String()
	item.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	err := svc.CreateDbItem(&item)
	if hp.ErrorCheck(res, err, 500) {
		return
	}
	res.WriteHeader(http.StatusCreated)
	hp.JSONResponse(res, item)
}

func getItem(res http.ResponseWriter, req *http.Request) {
	id, err := hp.CheckID(req.URL.Path)
	if hp.ErrorCheck(res, err, 400) {
		return
	}
	item, err := svc.GetDbItem(id)
	if hp.ErrorCheck(res, err, 500) {
		return
	}
	if item.ID == "" {
		http.Error(res, "item not found", 404)
		return
	}
	hp.JSONResponse(res, item)
}

func updateItem(res http.ResponseWriter, req *http.Request) {
	id, err := hp.CheckID(req.URL.Path)
	if hp.ErrorCheck(res, err, 400) {
		return
	}
	existItem, err := svc.GetDbItem(id)
	if hp.ErrorCheck(res, err, 500) {
		return
	}
	if existItem.ID == "" {
		http.Error(res, "item not found", 404)
		return
	}

	item := db.TodoItem{}
	if err := hp.JSONDecode(req.Body, &item); hp.ErrorCheck(res, err, 500) {
		return
	}
	if item.ID != "" {
		http.Error(res, "you try to update forbidden parameter id", 403)
		return
	}
	item.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	err = svc.UpdateDbItem(id, &item)
	if err != nil {
		hp.ErrorCheck(res, err, 500)
		return
	}
	res.WriteHeader(201)
	hp.JSONResponse(res, item)
}

func deleteItem(res http.ResponseWriter, req *http.Request) {
	id, err := hp.CheckID(req.URL.Path)
	if hp.ErrorCheck(res, err, 400) {
		return
	}
	err = svc.DeleteDbItem(id)
	if hp.ErrorCheck(res, err, 500) {
		return
	}
	res.WriteHeader(204)
}

func getAll(res http.ResponseWriter, req *http.Request) {
	itemsArr, err := svc.GetAllDbItems()
	if hp.ErrorCheck(res, err, 500) {
		return
	}
	hp.JSONResponse(res, itemsArr)
}

func main() {
	prepareEnvs()
	processFlags()
	svc.ConnectAndCreateTable(Envs)
	//set routes
	reHandler := new(regexpHandler)

	reHandler.HandleFunc("/$", "GET", homeController)
	reHandler.HandleFunc("/todo$", "POST", createNew)
	reHandler.HandleFunc("/todo/[(A-f0-9)]{24}$", "GET", getItem)
	reHandler.HandleFunc("/todo/[(A-f0-9)]{24}$", "PUT", updateItem)
	reHandler.HandleFunc("/todo/[(A-f0-9)]{24}$", "DELETE", deleteItem)
	reHandler.HandleFunc("/todos$", "GET", getAll)

	fmt.Printf("Listen port %v\n", PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%v", PORT), reHandler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
