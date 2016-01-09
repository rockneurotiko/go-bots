package downloader

import (
	"fmt"
	"log"
)

var WorkerQueue chan chan WorkRequest

func StartDispatcher(nworkers int) {
	WorkerQueue = make(chan chan WorkRequest, nworkers)
	workerschans := make([]Worker, 0)
	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Println("Starting worker", i+1)
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start()
		workerschans = append(workerschans, worker)
	}

	go func() {
		stopping := false
		for {
			select {
			case b := <-StopDispatcher:
				stopping = true
				fmt.Println("Stopping workers!")
				allans := make([]chan bool, 0)
				for i, worker := range workerschans {
					fmt.Printf("Stopping worker: %d\n", i+1)
					cans := worker.Stop()
					allans = append(allans, cans)
				}
				for _, cr := range allans {
					<-cr
				}
				b <- true
			case work := <-WorkQueue:
				// log.Println("Received work request")
				if stopping {
					work.Bot.Send(work.Id).Text("The bot are being closed, probably to deploy a new version, wait 5 minutes and try again :)").End()
					continue
				}
				go func() {
					worker := <-WorkerQueue
					// log.Println("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}
