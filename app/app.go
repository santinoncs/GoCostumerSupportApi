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
	AverageResponseTime float64
	QuestionsProcess    int
	timeCounter         time.Time
	TimeAnswered		float64
	IDs					[]ID
}

// ID : a new struct with just one field
type ID struct{
	ID	string
}

// QuestionStatus : here you tell us what Status of a question is
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

// QuestionPost : question method post
func (a *App) QuestionPost(priority int, question string) (Ack) {

	a.timeCounter = time.Now()

	
	q := newQuestion(priority, question)

	ack := Ack{
		ID:         q.ID,
		Success:    true,
		Message:    "",
	}

	questionStat := ID{
		ID: q.ID,
	}

	go func() {

		if priority == 1 {
			a.questionQueue <- q
			a.Status.QuestionsS()
			fmt.Println("Añado este ID al array", q.ID)
			a.Status.SetID(questionStat)
		}

	}()

	return ack

}

// GetNext : Return Question from questionQueue
func (a *App) GetNext() (Question, error) {

	var q Question

	select {
	case q, ok := <-a.questionQueue:
		if ok {
			return q, nil
		}
		return q, errors.New("channel closed")

	default:
		return q, errors.New("No Question available")
	}

}

func contains(s []ID, e string) bool {
    for _, a := range s {
        if a.ID == e {
            return true
        }
    }
    return false
}

// PostCsAnswer : used by customer support people to answer the question
func (a *App) PostCsAnswer(ID string, answer string) PostAnswerAck {

	var ans Answer

	ans.Answer = answer
	ans.ID = ID

	var ackpostanswered PostAnswerAck

	fmt.Println("This is the array ids", a.Status.IDs)


	if contains(a.Status.IDs, ID) == true {

		fmt.Println("Existe el ID", ID)

		go func() {

			a.answerQueue <- ans
			a.Status.QuestionsA()

		}()

		ackpostanswered = PostAnswerAck{
			Success: true,
			Message: "OK",
		}

	} else {
		fmt.Println("NOOOO existe el ID", ID)
		ackpostanswered = PostAnswerAck{
			Success: false,
			Message: "Error",
		}
	}

	return ackpostanswered



}

// GetQuestion : Get the status of the question with id param
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
 			a.Status.SetProcessed(elapsed.Seconds())
			return *s, nil
		}
	default:
		return *s, errors.New("Does not exist")
	}

	return *s,nil

}


// SetID : method SetID
func (s *Status ) SetID(e ID) {
	s.IDs = append(s.IDs,e)
	fmt.Printf("Value %v of IDs was added.\n", s.IDs)
}

// SetProcessed : method SetProcessed
func (s *Status ) SetProcessed(e float64) {
	s.TimeAnswered += e
}

// QuestionsA : method QuestionsProcessed
func (s *Status ) QuestionsA() {
	s.mutex.Lock()
	s.QuestionsAnswered ++
	s.mutex.Unlock()
}

// QuestionsS : method QuestionsSubmited
func (s *Status ) QuestionsS() {
	s.mutex.Lock()
	s.QuestionsSubmited ++
	s.mutex.Unlock()
}

// GetTotalStatus : this method will return s status struct
func (s *Status) GetTotalStatus() ( Status ) {

	s.AverageResponseTime = s.GetAverage()

	return *s
	
}

// GetAverage : method GetAverage
func (s *Status ) GetAverage() float64{
	var microsperprocess float64
	
	if s.QuestionsAnswered > 0 {
			microsperprocess = s.TimeAnswered / float64(s.QuestionsAnswered)
	} else {
			microsperprocess = 0
	}
	return microsperprocess
}