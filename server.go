package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// struct to be used in parsing the json data from the request body
type UrlBody struct {
	Url string
}

// @TODO find a way to not use this global variable
var pool *redis.Pool

func main() {
	// using redis pool to handle concurrent connections
	// as the redis Dial function is not safe to be used concurrently

	// initializing the redis pool
	pool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	// adding two handler functions
	// one for adding your own url
	// one for retrieving the url using "tinyUrl"
	mux := http.NewServeMux()
	mux.HandleFunc("/add", addHandler)
	mux.HandleFunc("/", retrieveHandler)

	// server running in port 8000
	fmt.Println("listening at port 8000...")
	http.ListenAndServe(":8000", mux)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// can handle only json post data
	// assumption that the user provides correct absolute url
	// @TODO check the given URL for correctness

	// check for POST HTTP method
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(405), 405)
		return
	}
	// read all the contents from the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// parsing the json data
	var u UrlBody
	json.Unmarshal(body, &u)
	url := u.Url

	// take only first 7 characters(assuming uniqueness)
	newTiny := calculateHash(url)[:7]

	conn := pool.Get()
	defer conn.Close()

	// checking if the url is already present
	// actually it is not required for our purpose
	// we will just overwrite the value with the same thing
	// but if we choose to add additional statistics it may be important
	// to check for already present URL
	// @TODO check if the URL is already present

	// assume that hash function has no collisions
	if _, err := conn.Do("SET", "myUrl:"+newTiny, url); err != nil {
		// no error is given by redis if the key is already present
		http.Error(w, http.StatusText(500), 500)
	}

	// if no error send the client the new tinyUrl
	fmt.Fprintf(w, "the new url is %s", newTiny)
}

func calculateHash(url string) string {
	// function to calculate the sha1 hash -- to obtain the unique tinyURL for every URL
	h := sha1.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}

func retrieveHandler(w http.ResponseWriter, r *http.Request) {

	// checking if the GET HTTP method is used
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// getting the tinyURL
	tinyUrl, err := getTinyUrl(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}

	conn := pool.Get()
	defer conn.Close()

	// obtaining the actual URL stored in redis
	url, err := redis.String(conn.Do("GET", "myUrl:"+tinyUrl))
	if err != nil {
		// checking if the error is because
		// the tinyURL is not found in redis

		// @TODO the error handling is not in idiomatic go.
		// can be converted into idiomatic code
		if err.Error() == "redigo: nil returned" {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		// handling other errors
		http.Error(w, http.StatusText(500), 500)
		log.Fatal(err)
		return
	}
	// redirecting the users to actual URL
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func getTinyUrl(tiny string) (string, error) {

	// @TODO the parsing is strictly adherent to the URL structure
	// so it breaks if the url structure is changed even slightly
	// a generic URL parsing function can be written
	arr := strings.Split(tiny, "/")
	if len(arr) != 2 {
		return "", errors.New("improper url")
	}
	return arr[1], nil
}
