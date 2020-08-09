package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

var workers = make(map[string]*Worker)
var workersLock = sync.RWMutex{}

type SavedWorker struct {
	Id                string   `json:"id"`
	SkypeToken        string   `json:"skypeToken"`
	Username          string   `json:"username"`
	Password          string   `json:"password"`
	HealthCheckThread string   `json:"healthCheckThread"`
	Managers          []string `json:"managers"`
	NsfwEnabledThreads []string `json:"nsfwEnabledThreads"`
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
			if savedWorker.Username != "" && savedWorker.Password != "" {
				token, err := Login(savedWorker.Username, savedWorker.Password)
				if err == nil {
					savedWorker.SkypeToken = token
				}
			}
			worker, err := NewWorker(savedWorker.SkypeToken, nil)
			if err != nil {
				log.Println("Fail to restore worker", savedWorker.Id)
			}
			worker.id = savedWorker.Id
			worker.healthCheckThread = savedWorker.HealthCheckThread
			worker.managers = savedWorker.Managers
			worker.nsfwEnabledThreads = savedWorker.NsfwEnabledThreads
			// TODO
			if savedWorker.Username != "" && savedWorker.Password != "" {
				worker.username = savedWorker.Username
				worker.password = savedWorker.Password
			}
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

func workerStatusCallback(worker *Worker) {
	if worker.autoRestart {
		//worker.Restart()
		go executeWithRetry(func() error {
			return worker.Restart()
		}, -1, 10*time.Second)
	}
}

// TODO
func executeWithRetry(function func() error, retry int, sleepInterval time.Duration) error {
	try := 0
	var execError error
	for ; retry < 0 || try < retry; {
		if execError = function(); execError != nil {
			time.Sleep(sleepInterval)
		} else {
			return nil
		}
		try ++
	}
	return execError
}

func AddWorker(w *Worker) {
	workersLock.Lock()
	defer workersLock.Unlock()
	workers[w.id] = w
	w.statusCallback = workerStatusCallback
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
