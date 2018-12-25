package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bou.ke/monkey"
	hp "github.com/user/todo-golang/helpers"
)

var reHandler = new(regexpHandler)

func freezeTime() *monkey.PatchGuard {
	wayback := time.Date(2018, time.September, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	return patch
}

func TestHomeController(t *testing.T) {
	//freeze time on server
	patch := freezeTime()
	defer patch.Unpatch()
	ALIVE = time.Now().Format("2006-01-02 15:04:05")

	testUpTimeCheck := upTimeCheck{
		true,
		ALIVE,
		ALIVE,
	}
	//setUp controller
	reHandler.HandleFunc("/$", "GET", homeController)

	//start test server
	ts := httptest.NewServer(reHandler)
	defer ts.Close()

	//send request to test server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Error("Health check failed", err)
	}
	//decode resp.body object
	respStruct := upTimeCheck{}
	hp.JSONDecode(resp.Body, &respStruct)

	//check for test errors
	if testUpTimeCheck != respStruct {
		t.Error("Health check failed, reponse object is not similar to sended")
	}
}
