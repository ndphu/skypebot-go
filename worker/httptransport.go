package worker

import (
	"encoding/json"
	"github.com/ndphu/skypebot-go/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func (w *Worker) setRequestHeaders(req *http.Request) {
	w.setDefaultHeaders(req)
	req.Header.Set("Authentication", "skypetoken="+w.skypeToken)
	req.Header.Set("RegistrationToken", w.registrationToken)
}
func (w *Worker) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")
	req.Header.Set("BehaviorOverride", "redirectAs404")
	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=Unknown; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
}

func (w *Worker) setMediaRequestHeaders(req *http.Request, transactionId string) {
	req.Header.Set("TransactionId", transactionId)
	req.Header.Set("Authorization", "skype_token "+w.skypeToken)
	req.Header.Set("RegistrationToken", w.registrationToken)
	req.Header.Set("X-Client-Version", "1418/8.62.0.83//")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")
}

func (w *Worker) executeHttpRequest(req *http.Request) (status int, headers http.Header, body []byte, err error) {
	log.Println("Making", req.Method, "request to", req.URL.String())
	client := http.Client{}
	resp, err := client.Do(req)
	status, headers, body, err = parseHttpResponse(resp, err)
	if status == 404 {
		se := model.SkypeError{}
		if err := json.Unmarshal(body, &se); err == nil {
			if se.ErrorCode == 752 {
				newUrl := headers.Get("Location")
				log.Println("Cloud location changed. Re-execute with new URL:", newUrl)
				if parse, err := url.Parse(newUrl); err == nil {
					w.baseUrl = parse.Scheme + "://" + parse.Host
					return -1, nil, nil, ErrorCloudLocationChanged
				}
			}
		}
	}
	return status, headers, body, err
}

func parseHttpResponse(resp *http.Response, err error) (int, http.Header, []byte, error) {
	if err != nil {
		log.Println("Fail to execute HTTP request", err)
		return -1, nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fail to read body", err)
		return resp.StatusCode, resp.Header, nil, err
	}
	return resp.StatusCode, resp.Header, respBody, nil
}

