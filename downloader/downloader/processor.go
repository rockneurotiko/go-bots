package downloader

import (
	"log"
	"net/http"

	"github.com/pmylund/go-cache"
	"github.com/rockneurotiko/go-tgbot"
)

type WorkRequest struct {
	Id            int
	Url           string
	Name          string
	Bot           tgbot.TgBot
	AnswerChannel chan WorkAnswer
}

func NewWorkRequest(msg tgbot.Message, url string, name string, bot tgbot.TgBot) (WorkRequest, chan WorkAnswer) {
	c := make(chan WorkAnswer)
	return WorkRequest{
		msg.Chat.ID,
		url,
		name,
		bot,
		c,
	}, c
}

type WorkAnswer struct {
	Result bool
}

var WorkQueue = make(chan WorkRequest, 10) // Simultaneous requests!

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	return Worker{ID: id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		log.Printf("Started worker %d", w.ID)
		for {
			w.WorkerQueue <- w.Work
			select {
			case work := <-w.Work:
				// log.Printf("Received work %v\n", work)
				res, err := http.Get(work.Url)
				if err != nil {
					log.Println(err)
					work.AnswerChannel <- WorkAnswer{false}
					continue
				}
				ans := work.Bot.Send(work.Id).Document(tgbot.ReaderSender{res.Body, work.Name}).End()
				res.Body.Close()
				if ans.Ok && ans.Result != nil && ans.Result.Document != nil {
					fileid := ans.Result.Document.FileID
					cacheids.Set(work.Url, fileid, cache.DefaultExpiration)
				}
				work.AnswerChannel <- WorkAnswer{true}
			case <-w.QuitChan:
				log.Printf("Stopping worker %d\n", w.ID)
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
