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
	QuestionStatus
}

// Answer : here you tell us what Answer is
type Answer struct {
	ID 	    string
	Answer	string
	Question
}

// Status : here you tell us what Status is
type Status struct {
	mutex               sync.Mutex
	QuestionsAnswered	int
	QuestionsSubmited   int
	AverageResponseTime int
	QuestionsProcess    int
	timeCounter         time.Time
	TimeAnswered		time.Duration
}

// QuestionStatus : here you tell us what Status is
type QuestionStatus struct {
	ID       			string
	Question
	Status				string
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
	Message    string
}

// PostAnswerAck : here you tell us what PostAnswerAck is
type PostAnswerAck struct {
	Success    bool
	Message string
}

// NewQuestionStatus : Constructor of status struct
func NewQuestionStatus() *QuestionStatus {
	return &QuestionStatus{}
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

	return q
}

// QuestionPost :
func (a *App) QuestionPost(priority int, question string) Ack {

	a.timeCounter = time.Now()

	q := newQuestion(priority, question)

	ack := Ack{
		ID:         q.ID,
		Success:    true,
		Message: "OK",
	}

	go func() {

		if priority == 1 {
			a.questionQueue <- q
			a.Status.QuestionsS()
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
	a.Answer.ID = ID

	ackpostanswered := PostAnswerAck{
		Success: true,
		Message: "OK",
	}


	go func() {

			a.answerQueue <- a.Answer
			a.Status.QuestionsA()

	}()

	return ackpostanswered

}

// GetQuestion :
func (a *App) GetQuestion(param string) ( QuestionStatus, error ) {

	s := NewQuestionStatus()

	select {
	case r, ok := <-a.answerQueue:
		if ok {
			s.Answer = r
			s.ID = r.ID
			s.Status = "answered"
			t := time.Now()
			elapsed := t.Sub(a.timeCounter)
 			a.Status.SetProcessed(elapsed)
			return *s, nil
		}
	default:
		fmt.Println("No value ready, moving on.")
		return *s, errors.New("Does not exist")
	}

	return *s,nil

}

// SetProcessed : method SetProcessed
func (s *Status ) SetProcessed(e time.Duration) {
	//s.mutex.Lock()
	s.TimeAnswered += e
	//s.mutex.Unlock()
}

// QuestionsA : method QuestionsProcessed
func (s *Status ) QuestionsA() {
	fmt.Println("increasing in 1 the answered processed")
	s.mutex.Lock()
	s.QuestionsAnswered ++
	s.mutex.Unlock()
}

// QuestionsS : method QuestionsSubmited
func (s *Status ) QuestionsS() {
	fmt.Println("increasing in 1 the questions processed")
	s.mutex.Lock()
	s.QuestionsSubmited ++
	s.mutex.Unlock()
}

func (s *Status) GetTotalStatus() ( Status ) {

	//s.averageResponseTime = s.GetAverage()
	return *s
	
}

// // GetAverage : method GetAverage
// func (s *Status ) GetAverage() int64{
// 	var microsperprocess int64
// 	micros := int64(s.TimeProcessed / time.Microsecond)
// 	if s.questionsAnswered > 0 {
// 			microsperprocess = micros / int64(s.questionsAnswered)
// 	} else {
// 			microsperprocess = 0
// 	}
// 	return microsperprocess
// }