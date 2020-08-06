package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

const defaultMediaBaseUrl = "https://api.asm.skype.com"

var ErrorEmptySkypeToken = errors.New("Empty Token")

type Endpoint struct {
	Id             string         `json:"id"`
	Type           string         `json:"type"`
	IsActive       bool           `json:"isActive"`
	ProductContext string         `json:"productContext"`
	Subscriptions  []Subscription `json:"subscriptions"`
}

type Subscription struct {
	Id                  int      `json:"id"`
	Type                string   `json:"type"`
	ChannelType         string   `json:"channelType"`
	ConversationType    int      `json:"conversationType"`
	EventChannel        string   `json:"eventChannel"`
	Template            string   `json:"template"`
	InterestedResources []string `json:"interestedResources"`
}

type Config struct {
	mediaBaseUrl      string       `json:"mediaBaseUrl"`
	messageBaseUrl    string       `json:"messageBaseUrl"`
	registrationToken string       `json:"registrationToken"`
	skypeToken        string       `json:"skypeToken"`
	endpoint          string       `json:"endpoint"`
	lock              sync.RWMutex `json:"_"`
}

var config *Config

func init() {
	skypeToken := loadTokenFromFile()
	if skypeToken != "" {
		messageBaseUrl, regToken, err := getCorrectMessageBaseUrl(skypeToken)
		if err != nil {
			config = &Config{
				mediaBaseUrl:      defaultMediaBaseUrl,
				messageBaseUrl:    "",
				skypeToken:        "",
				registrationToken: "",
			}
		} else {
			config = &Config{
				mediaBaseUrl:      defaultMediaBaseUrl,
				messageBaseUrl:    messageBaseUrl,
				skypeToken:        skypeToken,
				registrationToken: regToken,
			}
		}

	} else {
		config = &Config{
			mediaBaseUrl:      defaultMediaBaseUrl,
			messageBaseUrl:    "",
			skypeToken:        "",
			registrationToken: "",
		}
	}
}

func getCorrectMessageBaseUrl(skypeToken string) (string, string, error) {
	req, _ := http.NewRequest("GET", "https://client-s.gateway.messenger.live.com/v1/users/ME/properties", nil)
	setDefaultHeaders(req)
	req.Header.Set("Authentication", "skypetoken="+skypeToken)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
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
		return "", "", errors.New("skype token is expired")
	}
	parsedUrl, err := url.Parse(location)
	if err != nil {
		return "", "", err
	}
	endpoint := parsedUrl.Scheme + "://" + parsedUrl.Host
	registrationToken := resp.Header.Get("Set-RegistrationToken")

	log.Println("Endpoint", endpoint)

	return endpoint, registrationToken, nil
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")
	req.Header.Set("BehaviorOverride", "redirectAs404")
	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=Unknown; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
}

func loadEnpoints(messageBaseUrl, skypeToken, registrationToken string) ([]string, error) {
	ids := make([]string, 0)
	client := http.Client{}
	req, _ := http.NewRequest("GET", messageBaseUrl+"/v1/users/ME/endpoints", nil)
	req.Header.Set("Authentication", "skypetoken="+skypeToken)
	req.Header.Set("RegistrationToken", "registrationToken="+registrationToken)
	setDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(bodyData))

	endpoints := make([]Endpoint, 0)
	if err := json.Unmarshal(bodyData, &endpoints); err != nil {
		panic(err)
	}

	for _, e := range endpoints {
		ids = append(ids, e.Id)
	}

	return ids, nil
}

func Get() *Config {
	return config
}

func loadTokenFromFile() (string) {
	if data, err := ioutil.ReadFile("token"); err != nil {
		log.Println("Fail to load token from file")
		return ""
	} else {
		log.Println("Token loaded from file")
		return string(data)
	}
}

func (c *Config) ReloadWithSkypeToken(skypeToken string) (error) {
	if skypeToken != "" {
		endpoint, regToken, err := getCorrectMessageBaseUrl(skypeToken)
		if err != nil {
			return err
		}
		c.mediaBaseUrl = defaultMediaBaseUrl
		c.messageBaseUrl = endpoint
		c.skypeToken = skypeToken
		c.registrationToken = regToken
		return ioutil.WriteFile("token", []byte(skypeToken), 0755)
	}
	return ErrorEmptySkypeToken
}

func (c *Config) UpdateEndpointAndRegistrationToken(endpointId, registrationToken string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.endpoint = endpointId
	c.registrationToken = registrationToken
}

func (c *Config) SkypeToken() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.skypeToken
}

func (c *Config) RegistrationToken() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.registrationToken
}

func (c *Config) MediaBaseUrl() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.mediaBaseUrl
}

func (c *Config) MessageBaseUrl() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.messageBaseUrl
}

func (c *Config) CurrentEndpoint() string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.endpoint
}
