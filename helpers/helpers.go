package helpers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

//make own err message if id is not valid
type validateIDError struct {
	err string
}

func (e *validateIDError) Error() string {
	return e.err
}

//EnvCheck checks for set Env value
func EnvCheck(env string) string {
	val, ok := os.LookupEnv(env)
	if !ok {
		log.Fatalf("You don't set %v env value", env)
	}
	return val
}

// JSONResponse response with JsonObject if no errors
func JSONResponse(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json")

	payload, err := json.Marshal(data)
	if ErrorCheck(res, err, 500) {
		return
	}
	fmt.Fprintf(res, string(payload))
}

//JSONDecode decodes json from request.body
func JSONDecode(body io.ReadCloser, item interface{}) error {
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(item); err != nil {
		return err
	}
	return nil
}

//ErrorCheck checking for server or network errors
func ErrorCheck(res http.ResponseWriter, err error, status int) bool {
	if err != nil {
		http.Error(res, err.Error(), status)
		fmt.Println(err.Error())
		return true
	}
	return false
}

//CheckID looks for id in req path string, and returns it if valid
func CheckID(path string) (string, error) {
	re, err := regexp.Compile(`[(A-f0-9)]{24}$`)
	if err != nil {
		log.Fatal(err)
	}
	id := re.FindString(path)
	_, err = hex.DecodeString(id)
	if err != nil {
		err = &validateIDError{"invalid id"}
		return "", err
	}
	return id, nil
}
