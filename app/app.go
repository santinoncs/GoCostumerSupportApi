package app

import (
	"crypto/md5"
	"encoding/hex"
	_ "errors" // we would need this package
	 "fmt"    // we would need this package
	"strconv"
	"sync"
	"time"
)

// App : Basic struct
type App struct {
	Status			
	questionQueue []chan Question
	priority      int
	QuestionMap	  map[string]Question
	Answer		
}


// Status : here you tell us what Status is
type Status struct {
	mutex               sync.Mutex
	QuestionsAnswered	int
	QuestionsSubmited   int
	QuestionsQueued		int
	AverageResponseTime float64
	timeCounter         time.Time
	TimeAnswered		float64
	iDs					[]ID
	QueueLength		    map[int]int	// meant to store queue lenght 
}

// Answer : Answer struct
type Answer struct {
	ID 	    string
	Answer	string
}


// ID : a new struct with just one field
type ID struct{
	ID	string
}

// IncomingPostQuestion : here you tell us what IncomingPostQuestion is
type IncomingPostQuestion struct {
	Priority int    `json:"priority"`
	Question string `json:"question"`
}

// Question : here you tell us what Question is
type Question struct {
	ID       string
	Question string
	Priority int
	Status	 string
	Answer	string
}

// Ack : here you tell us what Ack is
type Ack struct {
	ID         string
	Success    bool
	Message    string
}

// PostAnswerAck : Response you give to client with each post question
type PostAnswerAck struct {
	Success    bool
	Message string
}


// NewApp : Constructor of App struct
func NewApp() *App {
	questionQueue := make([]chan Question, 3)
	for i := range questionQueue {
			questionQueue[i] = make(chan Question,100)
	}
	var length = make(map[int]int)
	var QuestionMap = make(map[string]Question)
	return &App{
		questionQueue: questionQueue,
		QuestionMap: QuestionMap,
		Status: Status{
			QueueLength: length,
		},
	}
}

// This function receives an string and generates a Unique ID
func generateHash(question string) ( string ) {

	now := time.Now().UnixNano()
	t := strconv.FormatInt(now, 10)
	s := question + t
	bs := md5.New()
	bs.Write([]byte(s))
	hash1 := hex.EncodeToString(bs.Sum(nil)[:3])

	return hash1

}

// Given a question string, we will return a Question struct with a random ID
func newQuestion(priority int, question string) Question {

	ID := generateHash(question)

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
			a.questionQueue[0] <- q
			a.Status.incrementLenght(1)
		}
		if priority == 2 {
			a.questionQueue[1] <- q
			a.Status.incrementLenght(2)
		}
		if priority == 3 {
			a.questionQueue[2] <- q
			a.Status.incrementLenght(3)
		}
		a.Status.QuestionsQ()
		a.Status.SetID(questionStat)
		q.Status = "queued"
		a.QuestionMap[q.ID] = q
		fmt.Println("This is queued", q)

	}()

	return ack

}

// GetNext : Return Question from questionQueue
func (a *App) GetNext() (Question) {


	
	for {
		select {
		case q := <-a.questionQueue[0]:
				q.Status = "in_progress"
				a.QuestionMap[q.ID] = q
				a.Status.QuestionsS()
				return q
		case q := <-a.questionQueue[1]:
				q.Status = "in_progress"
				a.QuestionMap[q.ID] = q
				a.Status.QuestionsS()
				return q
		case q := <-a.questionQueue[2]:
				q.Status = "in_progress"
				a.QuestionMap[q.ID] = q
				a.Status.QuestionsS()
				return q
		default:
				return Question{}
		}	
	}

	

}

// This function searchs and ID into an slice of IDs
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

	// var ans Answer

	if val, ok := a.QuestionMap[ID]; ok {
		fmt.Println(val)
		val.Status = "answered"
		val.Answer = answer
		a.QuestionMap[ID] = val
	}


	var ackpostanswered PostAnswerAck

	ackpostanswered = PostAnswerAck{
		Success: true,
		Message: "OK",
	}


	return ackpostanswered

}

// GetQuestion : Get the status of the question with id param
func (a *App) GetQuestion(param string) ( Question ) {


	if val, ok := a.QuestionMap[param]; ok {
		fmt.Println(val)
		return val
	}

	return Question{}


}

func (s *Status ) incrementLenght(priority int) {
	
	s.QueueLength[priority] ++

}

// SetID : method SetID
func (s *Status ) SetID(e ID) {
	s.iDs = append(s.iDs,e)
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

// QuestionsQ : method QuestionsQueued
func (s *Status ) QuestionsQ() {
	s.mutex.Lock()
	s.QuestionsQueued ++
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