package worker

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
)

func (w *Worker) PostImageToThread(target, objectId, fileName string, fileSize, width, height int) error {
	pmr := model.PostMessageRequest{
		MessageId:     "1" + utils.RandStringRunes(19),
		DisplayName:   w.skypeId,
		MessageType:   "RichText/UriObject",
		ContentType:   "text",
		ComposeTime:   utils.GetUTCNow(),
		Content:       getURIObjectContent(objectId, fileName, fileSize, width, height),
		AsmReferences: []string{objectId},
	}
	payload, err := json.Marshal(pmr)
	if err != nil {
		log.Println("Fail to unmarshal PostMessageRequest", err)
		return err
	}
	log.Println("payload", string(payload))
	req, _ := http.NewRequest("POST", w.baseUrl+"/v1/users/ME/conversations/"+target+"/messages", bytes.NewReader(payload))
	w.setRequestHeaders(req)
	status, headers, body, err := w.executeHttpRequest(req)
	if status != 201 {
		log.Println("Fail to post message", status)
		logHeaders(headers)
		log.Println(string(body))
		log.Println(err)
		return ErrorFailToPostMediaMessage
	}
	log.Println("Post message successfully")
	return nil
}

func (w *Worker) SendTextMessage(target, text string) (error) {
	pmr := model.PostMessageRequest{
		MessageId:   "1" + utils.RandStringRunes(16),
		DisplayName: w.skypeId,
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
	req, _ := http.NewRequest("POST", w.baseUrl+"/v1/users/ME/conversations/"+target+"/messages", bytes.NewReader(payload))
	w.setRequestHeaders(req)
	status, headers, body, err := w.executeHttpRequest(req)
	if err != nil {
		return err
	}
	if status != 201 {
		log.Println("Fail to send text message by error", status, string(body))
		logHeaders(headers)
		return ErrorFailToSendTextMessage
	}
	log.Println("Post message successfully")
	return nil
}

func (w *Worker) GetMessages(threadId string) ([]model.SkypeMessage, error) {
	threadMessages := make([]model.SkypeMessage, 0)

	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/conversations/"+threadId+"/messages?startTime=1&view=supportsExtendedHistory|msnp24Equivalent|supportsMessageProperties", nil)
	w.setRequestHeaders(req)
	_, _, resp, err := w.executeHttpRequest(req)
	if err != nil {
		log.Println("Fail to query message in thread", threadId)
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

func (w *Worker) GetAllTextMessagesWithLimitAndTimeout(target string, limit int) ([]model.SkypeMessage, error) {
	threadMessages := make([]model.SkypeMessage, 0)
	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/conversations/"+target+"/messages?pageSize=200&startTime=1&view=supportsExtendedHistory|msnp24Equivalent|supportsMessageProperties", nil)
	w.setRequestHeaders(req)
	//_, _, resp, err := w.executeHttpRequest(req)
	var resp []byte = nil
	if err := utils.ExecuteWithRetry(func() error {
		_, _, body, err := w.executeHttpRequest(req)
		if err != nil {
			return err
		}
		resp = body
		return nil
	}); err != nil {
		return threadMessages, err
	}

	respObj := model.GetMessagesResponse{}
	if err := json.Unmarshal(resp, &respObj); err != nil {
		log.Println("Fail to unmarshal response", err)
		return nil, err
	}

	for _, m := range respObj.Messages {
		if m.Type == "Message" {
			threadMessages = append(threadMessages, m)
			if limit > 0 && len(threadMessages) >= limit {
				return threadMessages, nil
			}
		}
	}

	if len(respObj.Messages) < 200 {
		return threadMessages, nil
	}

	syncState := respObj.MetaData.SyncState
	loop := 0
	for {
		loop = loop + 1
		_req2, _ := http.NewRequest("GET", syncState, nil)
		w.setRequestHeaders(_req2)
		var body []byte = nil
		utils.ExecuteWithRetry(func() error {
			_, _, b, err := w.executeHttpRequest(_req2)
			if err != nil {
				return err
			}
			body = b
			return nil
		})

		if err := json.Unmarshal(body, &respObj); err != nil {
			log.Println("Fail to unmarshal response", err)
			return nil, err
		}
		if len(respObj.Messages) == 0 {
			return threadMessages, nil
		}
		syncState = respObj.MetaData.SyncState
		for _, m := range respObj.Messages {
			if m.Type == "Message" {
				threadMessages = append(threadMessages, m)
				if limit > 0 && len(threadMessages) >= limit {
					return threadMessages, nil
				}
			}
		}
	}

	return threadMessages, nil
}

func (w *Worker) GetAllTextMessages(target string, limit int) ([]model.SkypeMessage, error) {
	threadMessages := make([]model.SkypeMessage, 0)
	req, _ := http.NewRequest("GET", w.baseUrl+"/v1/users/ME/conversations/"+target+"/messages", nil)
	q := req.URL.Query()
	q.Add("startTime", "2")
	q.Add("pageSize", "200")
	q.Add("view", "supportsExtendedHistory|msnp24Equivalent")
	req.URL.RawQuery = q.Encode()
	w.setRequestHeaders(req)
	//_, _, resp, err := w.executeHttpRequest(req)
	var resp []byte = nil
	if err := utils.ExecuteWithRetry(func() error {
		_, _, body, err := w.executeHttpRequest(req)
		if err != nil {
			return err
		}
		resp = body
		return nil
	}); err != nil {
		return threadMessages, err
	}

	respObj := model.GetMessagesResponse{}
	if err := json.Unmarshal(resp, &respObj); err != nil {
		log.Println("Fail to unmarshal response", err)
		return nil, err
	}

	for _, m := range respObj.Messages {
		if m.Type == "Message" && m.MessageType == "RichText" && m.SkypeEditedId == "" && m.Content != "" {
			threadMessages = append(threadMessages, m)
			if limit > 0 && len(threadMessages) >= limit {
				return threadMessages, nil
			}
		}
	}

	if len(respObj.Messages) < 200 {
		return threadMessages, nil
	}

	syncState := respObj.MetaData.SyncState
	loop := 0
	for {
		loop = loop + 1
		_req2, _ := http.NewRequest("GET", syncState, nil)
		w.setRequestHeaders(_req2)
		var body []byte = nil
		utils.ExecuteWithRetry(func() error {
			_, _, b, err := w.executeHttpRequest(_req2)
			if err != nil {
				return err
			}
			body = b
			return nil
		})

		if err := json.Unmarshal(body, &respObj); err != nil {
			log.Println("Fail to unmarshal response", err)
			return nil, err
		}
		if len(respObj.Messages) == 0 {
			return threadMessages, nil
		}
		syncState = respObj.MetaData.SyncState
		for _, m := range respObj.Messages {
			if m.Type == "Message" && m.MessageType == "RichText" && m.SkypeEditedId == "" && m.Content != "" {
				threadMessages = append(threadMessages, m)
				if limit > 0 && len(threadMessages) >= limit {
					return threadMessages, nil
				}
			}
		}
	}

	return threadMessages, nil
}

func (w *Worker) DeleteMessage(sm model.SkypeMessage) error {
	return utils.ExecuteWithRetry(func() error {
		req, _ := http.NewRequest("DELETE", w.baseUrl+"/v1/users/ME/conversations/"+sm.ConversationId+"/messages/"+sm.Id, nil)
		w.setRequestHeaders(req)
		_, _, _, err := w.executeHttpRequest(req)
		return err
	})
}

func getURIObjectContent(objectId, filename string, fileSize, width, height int) string {
	object := model.URIObject{
		Uri:          "https://api.asm.skype.com/v1/objects/" + objectId,
		UrlThumbnail: "https://api.asm.skype.com/v1/objects/" + objectId + "/views/imgt1_anim",
		Type:         "Picture.1",
		DocId:        objectId,
		Width:        width,
		Height:       height,
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
