package main

import "github.com/qingchengnus/cab-republic/db"
import "fmt"
import "github.com/gorilla/mux"
import "net/http"
import "log"
import "time"
import "github.com/gorilla/schema"
import "encoding/json"

var decoder = schema.NewDecoder()

func main() {
	err := database.InitializeDatabase()
	if err != nil {
		fmt.Println(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/users/signin", SignInHandler).Methods("POST")
	r.HandleFunc("/users", UpdatePreferenceHandler).Methods("PUT")
	r.HandleFunc("/intentions", CreateIntentionHandler).Methods("POST")
	// r.HandleFunc("/matchings", FindMatchHandler).Methods("GET")

	s := &http.Server{
		Addr:           ":8081",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

func SignInHandler(responseWriter http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()

	if err != nil {
		// Handle error
		fmt.Println(err)
	}

	decoder := schema.NewDecoder()
	// r.PostForm is a map of our POST form values
	c := new(credentials)
	err = decoder.Decode(c, request.PostForm)

	if err != nil {
		// Handle error
	}

	result, ageMin, ageMax, gender, accessToken := database.LogIn(request.FormValue("email"), request.FormValue("password"))

	if result {
		u := user{
			Age_min:      ageMin,
			Age_max:      ageMax,
			Gender:       gender,
			Access_token: accessToken,
		}
		resp := signInResponse{
			User: u,
		}

		b, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error:", err)
		}
		responseWriter.Write(b)
	} else {
		fmt.Println("Age min is ", ageMin)
		responseWriter.WriteHeader(401)
	}

}

type credentials struct {
	Email    string
	Password string
}

type signInResponse struct {
	User user
}

type user struct {
	Email        string
	Password     string
	Age_min      int
	Age_max      int
	Gender       int
	Access_token string
}

func UpdatePreferenceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	err := request.ParseForm()

	if err != nil {
		// Handle error
		fmt.Println(err)
	}
	// r.PostForm is a map of our POST form values
	u := new(updateInfo)
	err = decoder.Decode(u, request.PostForm)

	result := database.UpdateUser(u.Age_min, u.Age_max, u.Gender, token)

	if result {
		responseWriter.WriteHeader(200)
	} else {
		responseWriter.WriteHeader(401)
	}

}

type updateInfo struct {
	Age_min int
	Age_max int
	Gender  int
}

func CreateIntentionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	err := request.ParseForm()

	if err != nil {
		// Handle error
		fmt.Println(err)
	}
	// r.PostForm is a map of our POST form values
	i := new(intention)
	err = decoder.Decode(i, request.PostForm)

	result := database.CreateIntention(i.destination_latitude, i.destination_longitude, token)

	if result {
		responseWriter.WriteHeader(201)
	} else {
		responseWriter.WriteHeader(401)
	}
}

type intention struct {
	destination_latitude  float64
	destination_longitude float64
}
