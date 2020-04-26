package app

import (
	"crypto/md5"
	"encoding/hex"
	_ "errors"	// we would need this package
	"fmt"
	"strconv"
	"sync"
	"time"
)

// App : Basic struct
type App struct {
	Status
	questionQueue []chan Question
	priority      int
	Answer
	answerQueue   chan Answer
	QuestionStatus
}

// Answer : Answer struct
type Answer struct {
	ID 	    string
	Answer	string
}

// Status : here you tell us what Status is
type Status struct {
	mutex               sync.Mutex
	QuestionsAnswered	int
	QuestionsSubmited   int
	AverageResponseTime float64
	timeCounter         time.Time
	TimeAnswered		float64
	IDs					[]ID
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
	questionQueue := make([]chan Question, 3)
	for i := range questionQueue {
			questionQueue[i] = make(chan Question,100)
	}
	answerQueue := make(chan Answer, 100)
	return &App{
		questionQueue: questionQueue,
		answerQueue: answerQueue,
	}
}

func generateHash(question string) ( string ) {

	now := time.Now().UnixNano()
	t := strconv.FormatInt(now, 10)
	s := question + t
	bs := md5.New()
	bs.Write([]byte(s))
	hash1 := hex.EncodeToString(bs.Sum(nil)[:3])

	return hash1

}

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
		}
		if priority == 2 {
			a.questionQueue[1] <- q
		}
		if priority == 3 {
			a.questionQueue[2] <- q
		}
		a.Status.QuestionsS()
		fmt.Println("Añado este ID al array", q.ID)
		a.QuestionStatus.Status = "in_progress"
		a.Status.SetID(questionStat)


	}()

	return ack

}

// GetNext : Return Question from questionQueue
func (a *App) GetNext() (Question) {


	for {
		select {
		case q := <-a.questionQueue[0]:
				return q
		case q := <-a.questionQueue[1]:
				return q
		case q := <-a.questionQueue[2]:
				return q
		default:
				return Question{}
}
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
		//fmt.Println("NOOOO existe el ID", ID)
		ackpostanswered = PostAnswerAck{
			Success: false,
			Message: "Error. Question ID not found",
		}
	}

	return ackpostanswered

}

// GetQuestion : Get the status of the question with id param
func (a *App) GetQuestion(param string) ( QuestionStatus ) {

	select {
	case r, ok := <-a.answerQueue:
		if ok {
			fmt.Println("entering listen queue answer")
			a.QuestionStatus.Answer = r
			a.QuestionStatus.ID = r.ID
			a.QuestionStatus.Status = "answered"
			t := time.Now()
			elapsed := t.Sub(a.timeCounter)
 			a.Status.SetProcessed(elapsed.Seconds())
			return a.QuestionStatus
		} 
	default:
		if contains(a.Status.IDs, param) == true {
			// not yet proccessed
			a.QuestionStatus.Status = "in_progress"
		} else {
			// nothing in the queue
			a.QuestionStatus.Status = ""
			fmt.Println("Nothing to be done to this ID")
		}
		a.QuestionStatus.ID = param
		return a.QuestionStatus
	}

	return a.QuestionStatus

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