package config

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var Redirect = errors.New("redirect")

func Login() {
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return Redirect
	}
	req, _ := http.NewRequest("GET", "https://web.skype.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	redirect1 := resp.Header.Get("Location")
	log.Println(resp.Status, redirect1)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(body))

	parsed, err := url.Parse(redirect1)
	if err != nil {
		panic(err)
	}

	//for k := range parsed.Query() {
	//	log.Println(k, ":", parsed.Query().Get(k))
	//}

	clientId := parsed.Query().Get("client_id")
	partner := parsed.Query().Get("partner")
	redirectUri := parsed.Query().Get("redirect_uri")
	state := parsed.Query().Get("state")

	log.Println("clientId", clientId, "partner", partner, "redirectUri", redirectUri, "state", state)

	req2, _ := http.NewRequest("GET", redirect1, nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	resp2, _ := client.Do(req2)
	defer resp2.Body.Close()

	location2 := resp2.Header.Get("Location")
	log.Println(location2)


	parsed, _ = url.Parse(location2)
	for k := range parsed.Query() {
		log.Println(k, ":", parsed.Query().Get(k))
	}


	req3, _ := http.NewRequest("GET", location2, nil)
	resp3, _ := client.Do(req3)
	defer resp3.Body.Close()
	cookies := resp3.Header["Set-Cookie"]
	for _, cookie := range cookies {
		log.Println(cookie)
	}
	//cookie := resp3.Header.Get("Set-Cookie")
	//log.Println("Set cookie:", cookie)


}
