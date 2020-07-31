package poll

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/skype/chat"
	"github.com/ndphu/skypebot-go/skype/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type PollingResponse struct {
	ErrorCode     int            `json:"errorCode"`
	EventMessages []EventMessage `json:"eventMessages"`
}

type Resource struct {
	Id          string                `json:"id"`
	LastMessage model.ExistingMessage `json:"lastMessage"`
}

type EventMessage struct {
	Id           int                `json:"id"`
	ResourceLink string             `json:"resourceLink"`
	ResourceType string             `json:"resourceType"`
	Time         string             `json:"time"`
	Type         string             `json:"type"`
	Resource     NewMessageResource `json:"resource"`
}

type NewMessageResource struct {
	Type             string `json:"type"`
	From             string `json:"from"`
	ClientMessageId  string `json:"clientmessageid"`
	Content          string `json:"content"`
	ContentType      string `json:"contenttype"`
	ThreadTopic      string `json:"thread_topic"`
	ConversationLink string `json:"conversationLink"`
	Id               string `json:"id"`
}

var ErrorFailToCreateSubscription = errors.New("fail to create subscription")
var ErrorMissingEndpoint = errors.New("missing endpoint")

type SubscriptionRequest struct {
	ChannelType         string   `json:"channelType"`
	ConversationType    int      `json:"conversationType"`
	InterestedResources []string `json:"interestedResources"`
}

func StartPolling() error {
	return startPolling()
}

func startPolling() error {
	sigChan := make(chan error, 0)

	go func() {
		for {
			_, err := CreateEndpoint()
			if err != nil {
				log.Println("Fail to create endpoint")
				return
			}
			log.Println("Polling using endpoint:", config.Get().CurrentEndpoint())

			subscriptionId, err := CreateSubscriptionForNewMessage()
			if err != nil {
				log.Println("Fail to create subscription")
				return
			}

			go subscribe(subscriptionId, sigChan)
			go active(sigChan)
			go sendHeartbeatMessages(sigChan)
			rerror := <-sigChan
			log.Println("Some thing failing", rerror)
			log.Println("Sleep and retry polling")
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func subscribe(subscriptionId int, signChan chan error) {
	ackId := 0
	for {
		endpoint := config.Get().CurrentEndpoint()
		req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints/"+endpoint+
			"/subscriptions/"+strconv.Itoa(subscriptionId)+"/poll?ackId="+strconv.Itoa(ackId), nil)
		utils.SetRequestHeaders(req)
		utils.SetEndpointHeader(req)
		status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
		if err != nil {
			log.Println("Maybe long polling timeout. Retry...")
			continue
		}
		if status != 200 {
			log.Println("Fail to continue polling. Server return status", status, string(body))
			utils.LogHeaders(headers)
			signChan <- ErrorFailPolling
			return
		}
		pr := PollingResponse{}
		if err := json.Unmarshal(body, &pr); err != nil {
			continue
		}
		for _, em := range pr.EventMessages {
			log.Println(em.Id, em.ResourceType, em.Type)
			ackId = em.Id
			if em.ResourceType == "NewMessage" {
				processNewMessage(em)
			}
		}
	}
}

var ErrorFailActiveHeartBeat = errors.New("fail active heart beat")
var ErrorFailPolling = errors.New("fail polling thread")

func active(signChan chan error) {
	for {
		req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints/"+config.Get().CurrentEndpoint()+"/active", strings.NewReader("{\"timeout\":120}"))
		utils.SetRequestHeaders(req)
		utils.SetEndpointHeader(req)
		status, _, body, err := utils.ExecuteHttpRequestExtended(req)
		if err != nil {
			log.Println("Fail to post active", status, string(body))
			signChan <- ErrorFailActiveHeartBeat
			return
		}

		time.Sleep(15 * time.Second)
	}
}

func sendHeartbeatMessages(signChan chan error) {
	for {
		log.Println("Post active successfully")
		if err := chat.PostTextMessage("19:8052c0b5464f40aab38d73d641cbed11@thread.skype", time.Now().String()); err != nil {
			signChan <- ErrorFailActiveHeartBeat
			return
		}
		log.Println("Post heartbeat successfully")
		time.Sleep(1 * time.Minute)
	}

}

func GetEndpoints() ([]config.Endpoint, error) {
	req, _ := http.NewRequest("GET", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints", nil)
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

func CreateEndpoint() (*config.Endpoint, error) {
	req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints", strings.NewReader(
		`{"endpointFeatures":"Agent,Presence2015,MessageProperties,CustomUserProperties,Highlights,Casts,CortanaBot,ModernBots,AutoIdleForWebApi,secureThreads,notificationStream,InviteFree,SupportsReadReceipts"}`))
	utils.SetRequestHeaders(req)
	status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		log.Println("Fail to create endpoint")
		return nil, err
	}
	if status != 201 {
		log.Println("Fail to create endpoint", string(body))
		return nil, err
	}
	regToken := headers.Get("Set-RegistrationToken")
	endpointId := path.Base(headers.Get("Location"))
	decodedValue, err := url.QueryUnescape(endpointId)
	if err == nil {
		endpointId = decodedValue
	}
	log.Println("New endpoint id:", endpointId)

	config.Get().UpdateEndpointAndRegistrationToken(endpointId, regToken)

	endpoints, err := GetEndpoints()
	if err != nil {
		log.Println("Fail to get endpoint list")
		return nil, err
	}
	for _, e := range endpoints {
		if strings.Contains(endpointId, e.Id) {
			return &e, nil
		}
	}
	return nil, ErrorMissingEndpoint
}

func CreateSubscriptionForNewMessage() (int, error) {
	subReq := SubscriptionRequest{
		ChannelType:      "HttpLongPoll",
		ConversationType: 2047,
		InterestedResources: []string{
			"/v1/users/ME/conversations/ALL/properties",
			"/v1/users/ME/conversations/ALL/messages",
			"/v1/threads/ALL",
		},
	}
	payload, _ := json.Marshal(subReq)
	log.Println("Subscription Request", string(payload))
	req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints/"+config.Get().CurrentEndpoint()+"/subscriptions", bytes.NewReader(payload))
	utils.SetRequestHeaders(req)
	utils.SetEndpointHeader(req)
	status, headers, body, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		return -1, err
	}
	if status != 201 {
		log.Println("Fail to create subscription", status, string(body))
		utils.LogHeaders(headers)
		return -1, ErrorFailToCreateSubscription
	}
	log.Println("Response headers:")
	utils.LogHeaders(headers)

	return strconv.Atoi(path.Base(headers.Get("Location")))
}

func processNewMessage(evt EventMessage) {
	ProcessMessage(evt)
}
