package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

var workers = make(map[string]*Worker)
var workersLock = sync.RWMutex{}

type SavedWorker struct {
	Id         string `json:"id"`
	SkypeToken string `json:"skypeToken"`
}

func init() {
	if data, err := ioutil.ReadFile("workers.json"); err != nil {
		log.Println("Fail to read workers file. Continue...")
	} else {
		savedWorkers := make([]SavedWorker, 0)
		if err := json.Unmarshal(data, &savedWorkers); err != nil {
			log.Println("Fail to parse saved worker")
			return
		}
		for _, savedWorker := range savedWorkers {
			worker, err := NewWorker(savedWorker.SkypeToken, nil)
			if err != nil {
				log.Println("Fail to restore worker", savedWorker.Id)
			}
			worker.id = savedWorker.Id
			if err := worker.Start(); err != nil {
				log.Println("Fail to start worker", worker.id, "from saved file")
			} else {
				AddWorker(worker)
			}
		}
	}
}

func FindWorker(workerId string) (w *Worker) {
	workersLock.Lock()
	defer workersLock.Unlock()
	if w, exists := workers[workerId]; exists {
		return w
	}
	return nil
}

func AddWorker(w *Worker) {
	workersLock.Lock()
	defer workersLock.Unlock()
	workers[w.id] = w
}

func SaveWorkers() error {
	workersLock.Lock()
	defer workersLock.Unlock()
	savedWorkers := make([]SavedWorker, 0)
	for _, w := range workers {
		savedWorkers = append(savedWorkers, SavedWorker{
			Id:         w.id,
			SkypeToken: w.skypeToken,
		})
	}
	log.Println("Saving", len(savedWorkers), "to workers.json")
	if data, err := json.Marshal(savedWorkers); err != nil {
		return err
	} else {
		return ioutil.WriteFile("workers.json", data, 0755)
	}
}

func GetWorkers() ([]WorkerData) {
	workersLock.Lock()
	defer workersLock.Unlock()
	result := make([]WorkerData, 0)
	for _, w := range workers {
		result = append(result, w.Data())
	}
	return result
}
