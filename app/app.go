package app

import (
	"crypto/md5"
	"encoding/hex"
	"errors" // we would need this package
	_ "fmt"  // we would need this package
	"strconv"
	"sync"
	"time"
)

// App : Basic struct
type App struct {
	Status
	questionQueue []chan string
	priority      int
	QuestionDB
	Answer
}

// Status : here you tell us what Status is
type Status struct {
	mutex               sync.Mutex
	QuestionsAnswered   int
	QuestionsSubmited   int
	QuestionsQueued     int
	AverageResponseTime float64
	timeCounter         time.Time
	TimeAnswered        float64
	iDs                 []ID
	QueueLength         map[int]int // meant to store queue lenght
}

// Answer : Answer struct
type Answer struct {
	ID     string
	Answer string
}

// ID : a new struct with just one field
type ID struct {
	ID string
}

// IncomingPostQuestion : here you tell us what IncomingPostQuestion is
type IncomingPostQuestion struct {
	Priority int    `json:"priority"`
	Question string `json:"question"`
}

// Question : here you tell us what Question is
type Question struct {
	ID         string
	Question   string
	priority   int
	Status     string
	Answer     string
	AnswerTime time.Duration
}

// Ack : here you tell us what Ack is
type Ack struct {
	ID      string
	Success bool
	Message string
}

// PostAnswerAck : Response you give to client with each post question
type PostAnswerAck struct {
	Success bool
	Message string
}

// QuestionDB :
type QuestionDB struct {
	questionDBMap map[string]*Question
	mutex         sync.RWMutex
}

// SetQueued :  SetQueued
func (q *QuestionDB) SetQueued(ID string, question string) {

	q.mutex.Lock()
	q.questionDBMap[ID] = &Question{ID: ID, Status: "queued", Question: question}
	q.mutex.Unlock()

}

// SetInProgress :
func (q *QuestionDB) SetInProgress(ID string) {

	q.mutex.Lock()
	q.questionDBMap[ID].Status = "in_progress"
	q.mutex.Unlock()

}

// SetAnswered :  SetAnswered
func (q *QuestionDB) SetAnswered(ID string) {

	q.mutex.Lock()
	q.questionDBMap[ID].Status = "in_progress"
	q.mutex.Unlock()

}

// NewApp : Constructor of App struct
func NewApp() *App {
	questionQueue := make([]chan string, 3)
	for i := range questionQueue {
		questionQueue[i] = make(chan string, 100)
	}
	var length = make(map[int]int)
	var QuestionMap = make(map[string]*Question)
	return &App{
		questionQueue: questionQueue,
		QuestionDB: QuestionDB{
			questionDBMap: QuestionMap,
		},
		Status: Status{
			QueueLength: length,
		},
	}
}

// This function receives an string and generates a Unique ID
func generateHash(question string) string {

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

	q := Question{ID: ID, Question: question, priority: priority}

	return q
}

// QuestionPost : question method post
func (a *App) QuestionPost(priority int, question string) Ack {

	a.timeCounter = time.Now()

	q := newQuestion(priority, question)

	ack := Ack{
		ID:      q.ID,
		Success: true,
		Message: "",
	}

	questionStat := ID{
		ID: q.ID,
	}

	go func() {

		if priority == 1 {
			a.questionQueue[0] <- q.ID
			a.Status.incrementLenght(1)
		}
		if priority == 2 {
			a.questionQueue[1] <- q.ID
			a.Status.incrementLenght(2)
		}
		if priority == 3 {
			a.questionQueue[2] <- q.ID
			a.Status.incrementLenght(3)
		}
		a.Status.incrementQuestionsQueued()
		a.Status.SetID(questionStat)
		a.QuestionDB.SetQueued(q.ID, question)
	}()

	return ack

}

// GetNext : Return Question from questionQueue
func (a *App) GetNext() string {

	for {
		select {
		case qi := <-a.questionQueue[0]:
			a.QuestionDB.SetInProgress(qi)
			a.Status.incrementQuestionsSubmited()
			return qi
		case qi := <-a.questionQueue[1]:
			a.QuestionDB.SetInProgress(qi)
			a.Status.incrementQuestionsSubmited()
			return qi
		case qi := <-a.questionQueue[2]:
			a.QuestionDB.SetInProgress(qi)
			a.Status.incrementQuestionsSubmited()
			return qi
		default:
			return ""
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

	var ackpostanswered PostAnswerAck

	t := time.Now()
	elapsed := t.Sub(a.timeCounter)

	if val, ok := a.questionDBMap[ID]; ok {
		if a.questionDBMap[ID].Status == "in_progress" {
			a.QuestionDB.SetAnswered(ID)
			a.questionDBMap[ID].AnswerTime = elapsed
			val.Status = "answered"
			a.Status.incrementQuestionsAnswered()
			a.Status.SetProcessed(elapsed.Seconds())
			ackpostanswered = PostAnswerAck{
				Success: true,
				Message: "OK",
			}
		} else {
			ackpostanswered = PostAnswerAck{
				Success: false,
				Message: "Not in progress state",
			}
		}

	} else {
		ackpostanswered = PostAnswerAck{
			Success: false,
			Message: "Not exists ID",
		}
	}

	return ackpostanswered

}

// GetQuestion : Get the status of the question with id param
func (a *App) GetQuestion(param string) (Question, error) {

	defer a.QuestionDB.mutex.Unlock()

	a.QuestionDB.mutex.Lock()
	if val, ok := a.QuestionDB.questionDBMap[param]; ok {
		return *val, nil
	}

	return Question{}, errors.New("Item does not exist")

}

func (s *Status) incrementLenght(priority int) {

	s.QueueLength[priority]++

}

// SetID : method SetID
func (s *Status) SetID(e ID) {
	s.iDs = append(s.iDs, e)
}

// SetProcessed : method SetProcessed
func (s *Status) SetProcessed(e float64) {
	s.TimeAnswered += e
}

// incrementQuestionsAnswered : method incrementQuestionsAnswered
func (s *Status) incrementQuestionsAnswered() {
	s.mutex.Lock()
	s.QuestionsAnswered++
	s.mutex.Unlock()
}

// incrementQuestionsSubmited : method incrementQuestionsSubmited
func (s *Status) incrementQuestionsSubmited() {
	s.mutex.Lock()
	s.QuestionsSubmited++
	s.mutex.Unlock()
}

// incrementQuestionsQueued : method incrementQuestionsQueued
func (s *Status) incrementQuestionsQueued() {
	s.mutex.Lock()
	s.QuestionsQueued++
	s.mutex.Unlock()
}

// GetTotalStatus : this method will return s status struct
func (s *Status) GetTotalStatus() Status {

	s.AverageResponseTime = s.GetAverage()

	return *s

}

// GetAverage : method GetAverage
func (s *Status) GetAverage() float64 {
	var microsperprocess float64

	if s.QuestionsAnswered > 0 {
		microsperprocess = s.TimeAnswered / float64(s.QuestionsAnswered)
	} else {
		microsperprocess = 0
	}
	return microsperprocess
}
