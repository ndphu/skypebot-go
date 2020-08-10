package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

var ErrorFailToCreateEndpoint = errors.New("fail to create endpoint")
var ErrorSkypeTokenExpired = errors.New("skype token is expired")
var ErrorEmptySkypeToken = errors.New("empty skype token")
var ErrorFailToCreateSubscription = errors.New("fail to create subscription")
var ErrorFailPolling = errors.New("fail polling")
var ErrorFailToPostMediaMessage = errors.New("fail to post media message")
var ErrorLocationChanged = errors.New("location changed")
var ErrorFailToSendTextMessage = errors.New("fail to send text message")
var ErrorFailToInitWorker = errors.New("fail to init worker")

const defaultMediaBaseUrl = "https://api.asm.skype.com"

type Status string

const StatusStopped Status = "stopped"
const StatusStopping Status = "stopping"
const StatusRunning Status = "running"
const StatusStarting Status = "starting"

type WorkerData struct {
	BaseUrl      string `json:"baseUrl"`
	MediaBaseUrl string `json:"mediaBaseUrl"`
	Id           string `json:"id"`
	Status       Status `json:"status"`
	SkypeId      string `json:"skypeId"`
}

type Worker struct {
	stopRequest        chan bool
	endpoint           string
	skypeToken         string
	registrationToken  string
	subscriptionId     int
	eventCallback      EventCallback
	baseUrl            string
	mediaBaseUrl       string
	id                 string
	status             Status
	skypeId            string
	healthCheckThread  string
	autoRestart        bool
	statusCallback     StatusCallback
	username           string
	password           string
	managers           []string
	nsfwEnabledThreads []string
}

type EventCallback func(worker *Worker, event *model.MessageEvent)
type StatusCallback func(worker *Worker)

func NewWorker(skypeToken string, eventCallback EventCallback) (*Worker, error) {
	w := &Worker{
		skypeToken:        skypeToken,
		endpoint:          "",
		registrationToken: "",
		baseUrl:           "",
		mediaBaseUrl:      defaultMediaBaseUrl,
		stopRequest:       make(chan bool),
		id:                uuid.New().String(),
		status:            StatusStopped,
		eventCallback: func(worker *Worker, event *model.MessageEvent) {
			worker.ProcessMessage(event)
		},
		autoRestart: true,
	}
	if eventCallback != nil {
		w.eventCallback = eventCallback
	}
	return w, nil
}

func (w *Worker) Data() WorkerData {
	return WorkerData{
		Id:           w.id,
		Status:       w.status,
		BaseUrl:      w.baseUrl,
		MediaBaseUrl: w.mediaBaseUrl,
		SkypeId:      w.skypeId,
	}
}

func (w *Worker) Start() (error) {
	w.status = StatusStarting
	if err := utils.ExecuteWithRetryTimes(func() error {
		return w.loadBaseUrl()
	}, utils.RetryParams{Retry: 10, SleepInterval: 5*time.Second}); err != nil {
		log.Println("Fail to load base url", err.Error())
		return err
	}
	if err := w.createEndpoint(); err != nil {
		log.Println("Fail to create endpoint", err)
		if err == ErrorLocationChanged {
			log.Println("Location changed. Try to create endpoint again with new location.")
			if recreateError := w.createEndpoint(); recreateError != nil {
				log.Println("Fail to re-create endpoint")
				return recreateError
			}
		} else {
			return err
		}
	}
	w.startHealthCheck()
	return w.subscribe()
}

func (w *Worker) Stop() error {
	if w.status != StatusRunning {
		log.Println("Worker is not running.")
		return nil
	}
	w.status = StatusStopping
	log.Println("Stopping worker...")
	w.stopRequest <- true
	<-w.stopRequest
	return nil
}

func (w *Worker) createEndpoint() error {
	req, _ := http.NewRequest("POST", w.baseUrl+"/v1/users/ME/endpoints", strings.NewReader(
		`{"endpointFeatures":"Agent,Presence2015,MessageProperties,CustomUserProperties,Highlights,Casts,CortanaBot,ModernBots,AutoIdleForWebApi,secureThreads,notificationStream,InviteFree,SupportsReadReceipts"}`))
	w.setRequestHeaders(req)
	status, headers, body, err := w.executeHttpRequest(req)
	if err != nil {
		log.Println("Fail to create endpoint")
		return err
	}
	if status != 201 {
		log.Println("Fail to create endpoint", string(body))
		se := model.SkypeError{}
		if err := json.Unmarshal(body, &se); err != nil {
			log.Println("Fail to unmarshal Skype error", string(body))
			return err
		} else {
			if se.ErrorCode == 752 {
				// different cloud
				newLocation := headers.Get("Location")
				log.Println("Different cloud error. New location is:", newLocation)
				w.baseUrl = newLocation
				return ErrorLocationChanged
			}
		}
		return err
	}
	regToken := headers.Get("Set-RegistrationToken")
	endpointId := path.Base(headers.Get("Location"))
	decodedValue, err := url.QueryUnescape(endpointId)
	if err == nil {
		endpointId = decodedValue
	}
	log.Println("New endpoint id:", endpointId)
	w.endpoint = endpointId
	w.registrationToken = regToken

	endpoints, err := w.getEndpoints()
	if err != nil {
		log.Println("Fail to get endpoint list")
		return err
	}
	for _, e := range endpoints {
		if strings.Contains(endpointId, e.Id) {
			return nil
		}
	}
	return ErrorFailToCreateEndpoint
}

func (w *Worker) getEndpoints() ([]model.Endpoint, error) {
	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/endpoints", nil)
	w.setRequestHeaders(req)
	status, _, body, err := w.executeHttpRequest(req)
	if err != nil {
		return nil, err
	}
	endpoints := make([]model.Endpoint, 0)
	if err := json.Unmarshal(body, &endpoints); err != nil {
		log.Println("Status:", status, "Fail to unmarshal", string(body))
		return nil, err
	}
	return endpoints, nil
}

func (w *Worker) loadBaseUrl() (error) {
	jwt.Parse(w.skypeToken, func(token *jwt.Token) (interface{}, error) {
		claims := token.Claims.(jwt.MapClaims)
		w.skypeId, _ = claims["skypeid"].(string)
		return nil, nil
	})

	req, _ := http.NewRequest("GET", "https://client-s.gateway.messenger.live.com/v1/users/ME/properties", nil)
	w.setDefaultHeaders(req)
	req.Header.Set("Authentication", "skypetoken="+w.skypeToken)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if resp.StatusCode == 200 {
		location = "https://client-s.gateway.messenger.live.com"
	} else if resp.StatusCode == 401 {
		log.Println("Server return", resp.Status, )
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			log.Println("Respond body", string(body))
		}
		return ErrorSkypeTokenExpired
	} else if resp.StatusCode == 404 {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			se := model.SkypeError{}
			if err := json.Unmarshal(body, &se); err == nil {
				if se.ErrorCode == 752 {
					location = resp.Header.Get("Location")
					log.Println("Correct cloud:", location)
					parsedUrl, err := url.Parse(location)
					if err != nil {
						return err
					}
					w.baseUrl = parsedUrl.Scheme + "://" + parsedUrl.Host
					w.registrationToken = resp.Header.Get("Set-RegistrationToken")
					log.Println("Message base url:", w.baseUrl)
					log.Println("Loaded endpoint and registration token successfully")
					return nil
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}
	unknownResp, _ := ioutil.ReadAll(resp.Body)
	log.Println("Fail to init worker. Server response is unexpected:", resp.StatusCode, string(unknownResp))
	return ErrorFailToInitWorker
}

func (w *Worker) subscribe() error {
	subReq := model.SubscriptionRequest{
		ChannelType:      "HttpLongPoll",
		ConversationType: 2047,
		InterestedResources: []string{
			"/v1/users/ME/conversations/ALL/messages",
		},
	}
	payload, _ := json.Marshal(subReq)
	log.Println("Subscription Request", string(payload))
	req, _ := http.NewRequest("POST", w.baseUrl+"/v1/users/ME/endpoints/"+w.endpoint+"/subscriptions", bytes.NewReader(payload))
	w.setRequestHeaders(req)
	w.setEndpointHeader(req)
	status, headers, body, err := w.executeHttpRequest(req)
	if err != nil {
		return err
	}
	if status != 201 {
		log.Println("Fail to create subscription", status, string(body))
		logHeaders(headers)
		return ErrorFailToCreateSubscription
	}
	w.subscriptionId, _ = strconv.Atoi(path.Base(headers.Get("Location")))
	go w.startPolling()
	return nil
}

type HttpResult struct {
	resp *http.Response
	err  error
}

func (w *Worker) startPolling() {
	w.status = StatusRunning
	ackId := 0
	cx, cancel := context.WithCancel(context.Background())
	go func() {
		<-w.stopRequest
		cancel()
	}()

	defer func() {
		log.Println("Worker stopped successfully")
		w.status = StatusStopped
		w.stopRequest <- true
		if w.statusCallback != nil {
			go w.statusCallback(w)
		}
	}()
	client := http.Client{}
	resultChan := make(chan HttpResult)
	for {
		pollUrl := w.baseUrl + "/v1/users/ME/endpoints/" + w.endpoint + "/subscriptions/" + strconv.Itoa(w.subscriptionId) + "/poll"
		if ackId > 0 {
			pollUrl = pollUrl + "?ackId=" + strconv.Itoa(ackId)
		}
		req, _ := http.NewRequest("POST", pollUrl, nil)
		w.setRequestHeaders(req)
		w.setEndpointHeader(req)
		req.WithContext(cx)

		go func() {
			resp, err := client.Do(req)
			select {
			case <-cx.Done():
			default:
				resultChan <- HttpResult{resp, err}
			}
		}()

		select {
		case result := <-resultChan:
			status, headers, body, err := parseHttpResponse(result.resp, result.err)
			if err != nil {
				log.Println("Fail to make polling request", err)
				return
			}
			if status != 200 {
				log.Println("Fail to continue polling. Server return status", status, string(body))
				logHeaders(headers)
				return
			}
			pr := model.PollingResponse{}
			if err := json.Unmarshal(body, &pr); err != nil {
				continue
			}
			for _, em := range pr.Events {
				log.Println(em.Id, em.ResourceType, em.Type)
				ackId = em.Id
				if w.eventCallback != nil {
					w.eventCallback(w, &em)
				}
			}
		case <-cx.Done():
			log.Println("Stop http request received")
			return
		}
	}
}

func (w *Worker) setEndpointHeader(req *http.Request) {
	req.Header.Set("EndpointId", w.endpoint)
}
//
//func (w *Worker) Reload(skypeToken string) error {
//	if skypeToken != "" {
//		w.Stop()
//		w.skypeToken = skypeToken
//		return w.loadBaseUrl()
//	}
//	return ErrorEmptySkypeToken
//}

func (w *Worker) Restart() (error) {
	if err := w.Stop(); err != nil {
		log.Println("Fail to stop worker")
		return err
	}
	if w.username != "" && w.password != "" {
		token, err := Login(w.username, w.password)
		if err != nil {
			w.skypeToken = token
		}
	}
	if err := w.loadBaseUrl(); err != nil {
		log.Println("Fail to init worker")
		return err
	}
	return w.Start()
}

func (w*Worker) ShouldRelogin() bool {
	shouldRelogin := false
	jwt.Parse(w.skypeToken, func(token *jwt.Token) (interface{}, error) {
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
	return shouldRelogin
}