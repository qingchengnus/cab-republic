package main

import "github.com/qingchengnus/cab-republic/db"
import "fmt"
import "github.com/gorilla/mux"
import "net/http"
import "log"
import "time"
import "github.com/gorilla/schema"
import "encoding/json"
import "strings"

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
	r.HandleFunc("/matchings", FindMatchHandler).Methods("GET")
	r.HandleFunc("/matchings", DeleteMatchHandler).Methods("DELETE")
	r.HandleFunc("/matchings/poll", MatchPollHandler).Methods("GET")

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

	result, ageMin, ageMax, genderPreference, accessToken, userType := database.LogIn(request.FormValue("email"), request.FormValue("password"))

	if result {
		u := user{
			Age_min:           ageMin,
			Age_max:           ageMax,
			Gender_preference: genderPreference,
			Access_token:      accessToken,
			Type:              userType,
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
	Age_min           int
	Age_max           int
	Gender_preference int
	Access_token      string
	Type              int
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

	result := database.UpdateUser(u.Age_min, u.Age_max, u.Gender_preference, token)

	if result {
		responseWriter.WriteHeader(200)
	} else {
		responseWriter.WriteHeader(401)
	}

}

type updateInfo struct {
	Age_min           int
	Age_max           int
	Gender_preference int
}

func CreateIntentionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	err := request.ParseForm()
	fmt.Println(request)

	if err != nil {
		// Handle error
		fmt.Println(err)
	}
	// r.PostForm is a map of our POST form values
	i := new(intention)
	err = decoder.Decode(i, request.PostForm)
	fmt.Println(i.Destination_longitude)
	fmt.Println(i.Destination_latitude)
	result := database.CreateIntention(i.Destination_latitude, i.Destination_longitude, token)

	if result {
		responseWriter.WriteHeader(201)
	} else {
		responseWriter.WriteHeader(401)
	}
}

type intention struct {
	Destination_latitude  float64
	Destination_longitude float64
}

func FindMatchHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	q := request.URL.Query()
	emailsString := q["emails"][0]
	emails := strings.Split(emailsString, "-")

	result, email, point := database.FindMatch(emails, token)
	if result {
		m := match{
			Email:           email,
			Pickup_location: point,
		}

		b, err := json.Marshal(m)
		if err != nil {
			fmt.Println("error:", err)
		}
		responseWriter.Write(b)
	} else {
		responseWriter.WriteHeader(404)
	}
}

type match struct {
	Email           string
	Pickup_location string
}

func MatchPollHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	result, em, p := database.PollMatch(token)
	if result {
		e := MatchPollResponse{
			Email:           em,
			Pickup_location: p,
		}

		b, err := json.Marshal(e)
		if err == nil {
			responseWriter.Write(b)
			return
		}
	}
	responseWriter.WriteHeader(404)
}

type MatchPollResponse struct {
	Email           string
	Pickup_location string
}

func DeleteMatchHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := request.Header["Authorization"][0]
	result := database.DeleteMatch(token)
	if result {
		responseWriter.WriteHeader(200)
	} else {
		responseWriter.WriteHeader(404)
	}
}
