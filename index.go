package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type request struct {
	UrlList []string
}

type JsonResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func main() {

	// Init the mux router
	router := mux.NewRouter()

	router.HandleFunc("/health-check", healthCheck).Methods("GET")
	router.HandleFunc("/getImagesByUrl", downloadUsingChannel).Methods("GET")

	fmt.Println("Server at 8080")
	log.Fatal(http.ListenAndServe(":8080", router))

}

func healthCheck(w http.ResponseWriter, r *http.Request) {

	resp := JsonResponse{"success", "The application is up!"}
	json.NewEncoder(w).Encode(resp)

}

func downloadUsingChannel(w http.ResponseWriter, r *http.Request) {

	req, err := getBody(r)

	if err != nil {
		return
	}

	numRequests := len(req.UrlList)
	imageChan := make(chan []byte)

	for _, url := range req.UrlList {
		go getImage(url, imageChan)
	}

	count := 0

	for {
		data := <-imageChan
		count++
		ioutil.WriteFile("imageDump-"+strconv.Itoa(count)+".jpg", data, 0666)

		if count == numRequests {
			break
		}
	}

	resp := JsonResponse{"success", "All files have been downloaded"}
	json.NewEncoder(w).Encode(resp)

}

func getBody(r *http.Request) (request, error) {
	dec := json.NewDecoder(r.Body)

	var req request

	err := dec.Decode(&req)

	return req, err
}

func downloadImagesGoRoutines(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)

	var req request

	err := dec.Decode(&req)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(req.UrlList)

	for _, url := range req.UrlList {
		go downloadImageAndWrite(url)
	}

	resp := JsonResponse{"success", "Seems like it's working"}

	json.NewEncoder(w).Encode(resp)
}

func getImage(url string, c chan []byte) {

	resp, err := http.Get(url)

	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	resp.Body.Close()

	c <- data
}

func downloadUsingWaitGroup(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)

	var req request

	err := dec.Decode(&req)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(req.UrlList)

	// Adding multithreading
	var wg sync.WaitGroup
	wg.Add(len(req.UrlList))

	for _, url := range req.UrlList {
		go downloadImage(url, &wg)
	}

	fmt.Println("Waiting for goroutines to finish!")

	wg.Wait()

	fmt.Println("Completed all the threads")

	resp := JsonResponse{"success", "Seems like it's working"}

	json.NewEncoder(w).Encode(resp)

}

func downloadImage(url string, wg *sync.WaitGroup) {

	defer wg.Done()

	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return
	}

	err = ioutil.WriteFile("test_img"+fmt.Sprintf("%d", rand.Int())+".jpg", data, 0666)

	if err != nil {
		return
	}
}

func downloadImageAndWrite(url string) {

	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return
	}

	err = ioutil.WriteFile("test_img"+fmt.Sprintf("%d", rand.Int())+".jpg", data, 0666)

	if err != nil {
		return
	}
}
