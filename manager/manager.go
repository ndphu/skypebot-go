package manager

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/ndphu/skypebot-go/command"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"github.com/ndphu/skypebot-go/worker"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

var workers = make(map[string]*worker.Worker)
var workersLock = sync.RWMutex{}

type SavedWorker struct {
	Id                 string   `json:"id"`
	SkypeToken         string   `json:"skypeToken"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	HealthCheckThread  string   `json:"healthCheckThread"`
	Managers           []string `json:"managers"`
	NsfwEnabledThreads []string `json:"nsfwEnabledThreads"`
}

var WorkerEventCallback worker.EventCallback = func(worker *worker.Worker, evt *model.MessageEvent) {
	from := evt.GetFrom()
	threadId := evt.GetThreadId()

	log.Println("Processing message from", from, "on thread", threadId)
	if evt.Type == "EventMessage" && evt.ResourceType == "NewMessage" && evt.Resource.MessageType == "RichText" {
		if worker.IsMessageFromManager(evt) {
			//go w.processManageIM(evt)
			//go w.HandleAdminCommand(evt)
			command.HandleAdminCommand(worker, evt)
			return
		}

		if worker.IsDirectMention(evt) {
			go worker.ProcessMention(evt)
			return
		}

		if worker.IsDirectIM(evt) {
			go worker.ProcessDirectIM(evt)
			return
		}
	}
}

func Start() {
	if data, err := ioutil.ReadFile("workers.json"); err != nil {
		log.Println("Fail to read workers file. Continue...")
	} else {
		savedWorkers := make([]SavedWorker, 0)
		if err := json.Unmarshal(data, &savedWorkers); err != nil {
			log.Println("Fail to parse saved worker")
			return
		}
		for _, savedWorker := range savedWorkers {
			shouldRelogin := false
			jwt.Parse(savedWorker.SkypeToken, func(token *jwt.Token) (interface{}, error) {
				claims := token.Claims.(jwt.MapClaims)
				expiredAt := time.Unix(int64(claims["exp"].(float64)), 0)
				log.Println(expiredAt)
				remaining := expiredAt.Sub(time.Now())
				log.Println("token remaining time", remaining)
				if remaining < time.Hour {
					shouldRelogin = true
				}
				return nil, nil
			})
			log.Println("should re-login:", shouldRelogin)

			if shouldRelogin && savedWorker.Username != "" && savedWorker.Password != "" {
				token, err := Login(savedWorker.Username, savedWorker.Password)
				if err == nil {
					savedWorker.SkypeToken = token
				}
			}

			w, err := worker.NewWorker(savedWorker.SkypeToken, WorkerEventCallback)
			if err != nil {
				log.Println("Fail to restore worker", savedWorker.Id)
			}
			w.SetId(savedWorker.Id)
			w.SetHealthCheckThread(savedWorker.HealthCheckThread)
			w.SetManagers(savedWorker.Managers)
			//worker.nsfwEnabledThreads = savedWorker.NsfwEnabledThreads
			// TODO
			if savedWorker.Username != "" && savedWorker.Password != "" {
				w.SetUsername(savedWorker.Username)
				w.SetPassword(savedWorker.Password)
			}
			if err := w.Start(); err != nil {
				log.Println("Fail to start worker", w.GetId(), "from saved file")
			} else {
				AddWorker(w)
			}
		}
	}

	SaveWorkers()
}

func FindWorker(workerId string) (w *worker.Worker) {
	workersLock.Lock()
	defer workersLock.Unlock()
	if w, exists := workers[workerId]; exists {
		return w
	}
	return nil
}

func workerStatusCallback(worker *worker.Worker) {
	go utils.ExecuteWithRetryTimes(func() error {
		return worker.Restart()
	}, utils.RetryParams{
		Retry:         0,
		SleepInterval: 30 * time.Second,
	})
}

func AddWorker(w *worker.Worker) {
	workersLock.Lock()
	defer workersLock.Unlock()
	workers[w.GetId()] = w
	w.SetStatusCallback(workerStatusCallback)
}

func SaveWorkers() error {
	workersLock.Lock()
	defer workersLock.Unlock()
	savedWorkers := make([]SavedWorker, 0)
	for _, w := range workers {
		savedWorkers = append(savedWorkers, SavedWorker{
			Id:                 w.GetId(),
			SkypeToken:         w.GetSkypeToken(),
			Username:           w.GetUsername(),
			Password:           w.GetPassword(),
			HealthCheckThread:  w.GetHealthCheckThread(),
		})
	}
	log.Println("Saving", len(savedWorkers), "to workers.json")
	if data, err := json.Marshal(savedWorkers); err != nil {
		return err
	} else {
		return ioutil.WriteFile("workers.json", data, 0755)
	}
}

func GetWorkers() ([]worker.WorkerData) {
	workersLock.Lock()
	defer workersLock.Unlock()
	result := make([]worker.WorkerData, 0)
	for _, w := range workers {
		result = append(result, w.Data())
	}
	return result
}
