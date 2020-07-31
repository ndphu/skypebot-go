package poll

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ndphu/skypebot-go/config"
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
	_, err := CreateEndpoint()
	if err != nil {
		log.Println("Fail to create endpoint")
		return err
	}
	log.Println("Polling using endpoint:", config.Get().CurrentEndpoint())

	subscriptionId, err := CreateSubscriptionForNewMessage()
	if err != nil {
		log.Println("Fail to create subscription")
		return err
	}

	go subscribe(subscriptionId)
	go active()

	return nil
}

func subscribe(subscriptionId int) {
	ackId := 0
	for {
		endpoint := config.Get().CurrentEndpoint()
		req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints/"+endpoint+
			"/subscriptions/"+strconv.Itoa(subscriptionId)+"/poll?ackId="+strconv.Itoa(ackId), nil)
		utils.SetRequestHeaders(req)
		utils.SetEndpointHeader(req)
		status, _, body, err := utils.ExecuteHttpRequestExtended(req)
		if err != nil {
			log.Println("Maybe long polling timeout. Retry...")
			continue
		}
		if status != 200 {
			log.Println("Fail to continue polling. Server return status", status)
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
				processNewMessage(em, endpoint)
			}
		}
	}
}

func active() {
	for {
		req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/endpoints/"+config.Get().CurrentEndpoint()+"/active", strings.NewReader("{\"timeout\":120}"))
		utils.SetRequestHeaders(req)
		utils.SetEndpointHeader(req)
		_, _, _, err := utils.ExecuteHttpRequestExtended(req)
		if err != nil {
			log.Println("Fail to post active")
		} else {
			log.Println("Post active successfully")
		}
		time.Sleep(30 * time.Second)
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

func processNewMessage(evt EventMessage, endpoint string) {
	//
	//if
	//evt.Resource.From == "https://azwus1-client-s.gateway.messenger.live.com/v1/users/ME/contacts/8:dai.tran89" ||
	//	evt.Resource.From == "https://azwus1-client-s.gateway.messenger.live.com/v1/users/ME/contacts/8:letuankhang" ||
	//	evt.Resource.From == "https://azwus1-client-s.gateway.messenger.live.com/v1/users/ME/contacts/8:ngdacphu" {
	//	parts := strings.Split(evt.ResourceLink, "/")
	//	msgId, _ := strconv.Atoi(parts[9])
	//	chat.ReactMessage(parts[7], msgId, "poop", endpoint)
	//
	//}
	ProcessMessage(evt)
}
