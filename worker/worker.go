package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

var ErrorFailToCreateEndpoint = errors.New("fail to create endpoint")
var ErrorSkypeTokenExpired = errors.New("skype token is expired")
var ErrorFailToCreateSubscription = errors.New("fail to create subscription")
var ErrorFailPolling = errors.New("fail polling")

const defaultMediaBaseUrl = "https://api.asm.skype.com"

type PollingWorker struct {
	stopRequest       chan bool
	sigChan           chan error
	endpoint          string
	skypeToken        string
	registrationToken string
	baseUrl           string
	mediaBaseUrl      string
	subscriptionId    int
	eventCallback     EventCallback
}

type EventCallback func(event *EventMessage)

func NewWorker(skypeToken string, eventCallback EventCallback) (*PollingWorker, error) {
	worker := &PollingWorker{
		skypeToken:        skypeToken,
		endpoint:          "",
		registrationToken: "",
		baseUrl:           "",
		mediaBaseUrl:      defaultMediaBaseUrl,
		stopRequest:       make(chan bool),
		sigChan:           make(chan error),
		eventCallback:     eventCallback,
	}
	// set message base URL and registration token
	if err := worker.getCorrectMessageBaseUrl(); err != nil {
		return nil, err
	}
	// create endpoint

	return worker, nil
}

func (w *PollingWorker) Start() (error) {
	if err := w.createEndpoint(); err != nil {
		return err
	}
	return w.start()
}

func (w *PollingWorker) Stop() {
	log.Println("Stopping worker...")
	w.stopRequest <- true
	<-w.stopRequest
	log.Println("Working stopped successfully")
}

func (w *PollingWorker) createEndpoint() error {
	req, _ := http.NewRequest("POST", w.baseUrl+"/v1/users/ME/endpoints", strings.NewReader(
		`{"endpointFeatures":"Agent,Presence2015,MessageProperties,CustomUserProperties,Highlights,Casts,CortanaBot,ModernBots,AutoIdleForWebApi,secureThreads,notificationStream,InviteFree,SupportsReadReceipts"}`))
	w.setRequestHeaders(req)
	status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		log.Println("Fail to create endpoint")
		return err
	}
	if status != 201 {
		log.Println("Fail to create endpoint", string(body))
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

func (w *PollingWorker) getEndpoints() ([]config.Endpoint, error) {
	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/endpoints", nil)
	utils.SetRequestHeaders(req)
	status, _, body, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		return nil, err
	}
	endpoints := make([]config.Endpoint, 0)
	if err := json.Unmarshal(body, &endpoints); err != nil {
		log.Println("Status:", status, "Fail to unmarshal", string(body))
		return nil, err
	}
	return endpoints, nil
}

func (w *PollingWorker) getCorrectMessageBaseUrl() (error) {
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
	}
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

func (w *PollingWorker) subscribe() error {
	subReq := SubscriptionRequest{
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
	status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		return err
	}
	if status != 201 {
		log.Println("Fail to create subscription", status, string(body))
		utils.LogHeaders(headers)
		return ErrorFailToCreateSubscription
	}
	w.subscriptionId, _ = strconv.Atoi(path.Base(headers.Get("Location")))
	go w.startPolling()
	return nil
}

func (w *PollingWorker) start() error {
	err := w.createEndpoint()
	if err != nil {
		log.Println("Fail to create endpoint", err)
		return err
	}
	return w.subscribe()
}

func (w *PollingWorker) startPolling() {
	ackId := 0
	for {
		pollUrl := w.baseUrl + "/v1/users/ME/endpoints/" + w.endpoint + "/subscriptions/" + strconv.Itoa(w.subscriptionId) + "/poll"
		if ackId > 0 {
			pollUrl = pollUrl + "?ackId=" + strconv.Itoa(ackId)
		}
		req, _ := http.NewRequest("POST", pollUrl, nil)
		w.setRequestHeaders(req)
		w.setEndpointHeader(req)
		status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
		if err != nil {
			log.Println("Maybe long polling timeout. Retry...")
			continue
		}
		if status != 200 {
			log.Println("Fail to continue polling. Server return status", status, string(body))
			utils.LogHeaders(headers)
			w.sigChan <- ErrorFailPolling
			break
		}
		pr := PollingResponse{}
		if err := json.Unmarshal(body, &pr); err != nil {
			continue
		}
		for _, em := range pr.EventMessages {
			log.Println(em.Id, em.ResourceType, em.Type)
			ackId = em.Id
			if w.eventCallback != nil {
				w.eventCallback(&em)
			}
		}
	}
}

func parseInfo(evt EventMessage) (string, string) {
	threadId := path.Base(evt.Resource.ConversationLink)
	from := path.Base(evt.Resource.From)
	return threadId, from
}

func (w *PollingWorker) setEndpointHeader(req *http.Request) {
	req.Header.Set("EndpointId", w.endpoint)
}
