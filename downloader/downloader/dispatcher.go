package downloader

import "log"

var WorkerQueue chan chan WorkRequest

func StartDispatcher(nworkers int) {
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Println("Starting worker", i+1)
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				// log.Println("Received work request")
				go func() {
					worker := <-WorkerQueue
					// log.Println("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()

}
