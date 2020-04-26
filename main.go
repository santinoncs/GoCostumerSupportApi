package main

import (
	"fmt"
	app "github.com/santinoncs/GoCostumerSupportApi/app"
	//"sync"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)



var application *app.App

func main() {

	application = app.NewApp()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	err := http.ListenAndServe(":8080", nil)

	fmt.Println(err)

}

func handler(w http.ResponseWriter, r *http.Request) {

	var responseAck app.Ack
	var responseQuestion app.Question
	var responseAnswer app.PostAnswerAck
	var content app.IncomingPostQuestion
	var postAnswer app.Answer
	var responseTotalStatus app.Status
	var responseGetQuestion app.QuestionStatus

	if r.URL.Path == "/api/question/post" {

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&content); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			responseAck = app.Ack{
				Success: false,
				ID: "",
				Message: "error",
			}
		} else {
			responseAck = application.QuestionPost(content.Priority, content.Question)
		}

		
		responseJSON, _ := json.Marshal(responseAck)
		fmt.Fprintf(w, "Response: %s\n", responseJSON)

	}

	if r.URL.Path == "/api/question/get_next" {

		responseQuestion = application.GetNext()
		responseJSON, _ := json.Marshal(responseQuestion)
		fmt.Fprintf(w, "Response: %s\n", responseJSON)

	}

	if r.URL.Path == "/api/status" {

		responseTotalStatus = application.GetTotalStatus()
		responseJSON, _ := json.Marshal(responseTotalStatus)
		fmt.Fprintf(w, "Response: %s\n", responseJSON)

	}

	if r.URL.Path == "/api/question/answer" {

		err := json.NewDecoder(r.Body).Decode(&postAnswer)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		responseAnswer = application.PostCsAnswer(postAnswer.ID, postAnswer.Answer)
		responseJSON, _ := json.Marshal(responseAnswer)
		fmt.Fprintf(w, "Response: %s\n", responseJSON)

	}

	if r.URL.Path == "/api/question/status" {

		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
				log.Fatal(err)
		}

		param := params.Get("question_id")

		responseGetQuestion = application.GetQuestion(param)
		
		responseJSON, _ := json.Marshal(responseGetQuestion)
		fmt.Fprintf(w, "Response: %s\n", responseJSON)

	}


}
