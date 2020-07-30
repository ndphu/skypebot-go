package chat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/skype/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
	"time"
)

func PostImageToThread(target, objectId, fileName string, fileSize int) error {
	pmr := model.PostMessageRequest{
		MessageId:     "1" + utils.RandStringRunes(19),
		DisplayName:   "/dev/null",
		MessageType:   "RichText/UriObject",
		ContentType:   "text",
		ComposeTime:   utils.GetUTCNow(),
		Content:       getURIObjectContent(objectId, fileName, fileSize),
		AsmReferences: []string{objectId},
	}
	payload, err := json.Marshal(pmr)
	if err != nil {
		log.Println("Fail to unmarshal PostMessageRequest", err)
		return err
	}
	req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/conversations/"+target+"/messages", bytes.NewReader(payload))
	req.Header.Set("Authorization", "skype_token "+config.Get().SkypeToken())
	req.Header.Set("X-Client-Version", "1418/8.62.0.83//")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")

	utils.ExecuteHttpRequestExtended(req)
	log.Println("Post message successfully")
	return nil
}

func PostTextMessage(target, text string) (error) {
	pmr := model.PostMessageRequest{
		MessageId:   "1" + utils.RandStringRunes(19),
		DisplayName: "/dev/null",
		MessageType: "RichText",
		ContentType: "text",
		ComposeTime: utils.GetUTCNow(),
		Content:     text,
	}
	payload, err := json.Marshal(pmr)
	if err != nil {
		log.Println("Fail to unmarshal PostMessageRequest", err)
		return err
	}
	req, _ := http.NewRequest("POST", config.Get().MessageBaseUrl()+"/v1/users/ME/conversations/"+target+"/messages", bytes.NewReader(payload))
	utils.SetRequestHeaders(req)
	utils.ExecuteHttpRequestExtended(req)
	log.Println("Post message successfully")
	return nil
}

func GetMessages(target string) ([]model.ExistingMessage, error) {
	threadMessages := make([]model.ExistingMessage, 0)

	req, _ := http.NewRequest("GET", config.Get().MessageBaseUrl()+"/v1/users/ME/conversations/"+target+"/messages?startTime=1&view=supportsExtendedHistory|msnp24Equivalent|supportsMessageProperties", nil)
	utils.SetRequestHeaders(req)
	_, _, resp, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		log.Println("Fail to query message in thread", target)
		return nil, err
	}
	respObj := model.GetMessagesResponse{}
	if err := json.Unmarshal(resp, &respObj); err != nil {
		log.Println("Fail to unmarshal response", err)
		return nil, err
	}

	for _, m := range respObj.Messages {
		threadMessages = append(threadMessages, m)
	}

	return threadMessages, nil
}

func GetAllTextMessagesWithLimitAndTimeout(target string, limit int) ([]model.ExistingMessage, error) {
	threadMessages := make([]model.ExistingMessage, 0)

	req, _ := http.NewRequest("GET", config.Get().MessageBaseUrl()+"/v1/users/ME/conversations/"+target+"/messages?pageSize=200&startTime=1&view=supportsExtendedHistory|msnp24Equivalent|supportsMessageProperties", nil)
	utils.SetRequestHeaders(req)
	_, _, resp, err := utils.ExecuteHttpRequestExtended(req)
	if err != nil {
		log.Println("Fail to query message in thread", target)
		return nil, err
	}
	respObj := model.GetMessagesResponse{}
	if err := json.Unmarshal(resp, &respObj); err != nil {
		log.Println("Fail to unmarshal response", err)
		return nil, err
	}

	for _, m := range respObj.Messages {
		if m.Type == "Message" {
			threadMessages = append(threadMessages, m)
			if len(threadMessages) >= limit {
				return threadMessages, nil
			}
		}
	}

	syncState := respObj.MetaData.SyncState
	loop := 0
	for {
		loop = loop + 1
		log.Println("looping", loop, syncState)
		_req2, _ := http.NewRequest("GET", syncState, nil)
		utils.SetRequestHeaders(_req2)
		_, _, resp2, err := utils.ExecuteHttpRequestExtended(_req2)
		if err != nil {
			if err == utils.ErrorLimitRequestExceeded {
				time.Sleep(15 * time.Second)
				continue
			} else {
				log.Println("Fail to query message in thread", target)
				return threadMessages, nil
			}
		}
		if err := json.Unmarshal(resp2, &respObj); err != nil {
			log.Println("Fail to unmarshal response", err)
			return nil, err
		}
		syncState = respObj.MetaData.SyncState
		if len(respObj.Messages) == 0 {
			return threadMessages, nil
		}
		for _, m := range respObj.Messages {
			if m.Type == "Message" {
				threadMessages = append(threadMessages, m)
				if len(threadMessages) >= limit {
					return threadMessages, nil
				}
			}
		}
	}

	return threadMessages, nil
}

func getURIObjectContent(objectId, filename string, fileSize int) string {
	object := model.URIObject{
		Uri:          "https://api.asm.skype.com/v1/objects/" + objectId,
		UrlThumbnail: "https://api.asm.skype.com/v1/objects/" + objectId + "/views/imgt1_anim",
		Type:         "Picture.1",
		DocId:        objectId,
		Width:        0,
		Height:       0,
		Text:         "To view this shared photo, go to:",
		ViewLink: model.ViewLink{
			Href: "https://login.skype.com/login/sso?go=xmmfallback?pic=" + objectId,
			Link: "https://login.skype.com/login/sso?go=xmmfallback?pic=" + objectId,
		},
		OriginalName: model.OriginalName{
			Name: filename,
		},
		FileSize: model.FileSize{
			Size: fileSize,
		},
		Meta: model.Meta{
			Type:         "photo",
			OriginalName: filename,
		},
	}
	xmlData, err := xml.Marshal(object)
	if err != nil {
		panic(err)
	}
	return string(xmlData)
}
