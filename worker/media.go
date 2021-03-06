package worker

import (
	"bytes"
	"encoding/json"
	"github.com/ndphu/skypebot-go/model"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func (w *Worker) CreateObject(target, filename, transactionId string) (string, error) {
	permissions := make(map[string][]string)
	permissions[target] = []string{"read"}
	body := model.CreateObjectRequest{
		Type:        "pish/image",
		Filename:    filename,
		Permissions: permissions,
	}
	bodyPayload, err := json.Marshal(body)
	if err != nil {
		log.Println("Fail to marshal body", body)
		return "", err
	}

	req, err := http.NewRequest("POST", w.mediaBaseUrl+"/v1/objects", bytes.NewReader(bodyPayload))
	w.setMediaRequestHeaders(req, transactionId)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Fail to execute HTTP request", err)
		return "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fail to read body", err)
		return "", err
	}
	cor := model.CreateObjectResponse{}
	if err := json.Unmarshal(respBody, &cor); err != nil {

		log.Println("Fail to unmarshal response", resp.Status, string(respBody))
		return "", err
	}

	return cor.Id, nil
}

func (w *Worker) UploadObject(objectId, transactionId string, payload []byte) error {
	log.Println("Uploading object", objectId, "with data size", strconv.Itoa(len(payload)))
	req, err := http.NewRequest("PUT", w.mediaBaseUrl+"/v1/objects/"+objectId+"/content/imgpsh", bytes.NewReader(payload))
	w.setMediaRequestHeaders(req, transactionId)
	req.Header.Set("Content-Type", "application")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Fail to execute HTTP request", err)
		return err
	}
	defer resp.Body.Close()
	log.Println("Uploaded object", objectId, "successfully")
	return nil
}
