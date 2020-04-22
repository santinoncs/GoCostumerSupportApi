package app

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// App : here you tell us what App is
type App struct {
	Status
	questionQueue chan Question
	priority      int
	Answer
	answerQueue   chan Answer
}

// Answer : here you tell us what Answer is
type Answer struct {
	ID 	    string
	Answer	string
}

// Status : here you tell us what Status is
type Status struct {
	Question
	Status				string
	TimeAnswered        time.Duration
	mutex               sync.Mutex
	Answer
}

// Question : here you tell us what Question is
type Question struct {
	ID       string
	Question string
	Priority int
}

// Ack : here you tell us what Ack is
type Ack struct {
	ID         string
	Success    bool
	AckMessage string
}

// PostAnswerAck : here you tell us what PostAnswerAck is
type PostAnswerAck struct {
	Success    bool
	AckMessage string
}

// NewStatus : Constructor of status struct
func NewStatus() *Status {
	return &Status{}
}

// NewApp : here you tell us what NewApp is
func NewApp() *App {
	questionQueue := make(chan Question, 100)
	answerQueue := make(chan Answer, 100)
	return &App{
		questionQueue: questionQueue,
		answerQueue: answerQueue,
	}
}

func newQuestion(priority int, question string) Question {

	fmt.Println("accessing new Question")

	now := time.Now().UnixNano()

	t := strconv.FormatInt(now, 10)

	s := question + t

	bs := md5.New()

	bs.Write([]byte(s))

	hash1 := hex.EncodeToString(bs.Sum(nil)[:3])

	ID := hash1 // call random generation number here

	q := Question{ID: ID, Question: question, Priority: priority}

	fmt.Println("creating question:", q)

	return q
}

// QuestionPost :
func (a *App) QuestionPost(priority int, question string) Ack {

	q := newQuestion(priority, question)

	ack := Ack{
		ID:         q.ID,
		Success:    true,
		AckMessage: "delivered",
	}

	go func() {

		if priority == 1 {
			a.questionQueue <- q
		}

	}()

	return ack

}

// GetNext :
func (a *App) GetNext() (Question, error) {

	var q Question

	select {
	case q, ok := <-a.questionQueue:
		if ok {
			fmt.Printf("Value %v was read.\n", q)
			return q, nil
		}
		fmt.Println("Channel closed!")
		return q, errors.New("channel closed")

	default:
		fmt.Println("No value ready, moving on.")
		return q, errors.New("Item does not exist")
	}

}

// PostCsAnswer :
func (a *App) PostCsAnswer(ID string, answer string) PostAnswerAck {

	a.Answer.Answer = answer

	ackpostanswered := PostAnswerAck{
		Success: true,
		AckMessage: "",
	}

	fmt.Println("Customer Service posting this answer:", a.Answer.Answer )

	go func() {

			a.answerQueue <- a.Answer


	}()

	return ackpostanswered

}

// GetStatus :
func (a *App) GetStatus(param string) ( Status, error ) {

	s := NewStatus() 

	fmt.Println(a.answerQueue)

	select {
	case r, ok := <-a.answerQueue:
		if ok {
			fmt.Printf("Value %v was read.\n", r)
			s.Answer = r
			return *s, nil
		}
	default:
		fmt.Println("No value ready, moving on.")
		return *s, errors.New("Item does not exist")
	}

	return *s,nil

}