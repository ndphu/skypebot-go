package manager

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var Redirect = errors.New("redirect")

func Login(username, password string) (string, error) {
	return doLogin(username, password)
}

func doLogin(username, password string) (string, error) {
	client := http.Client{}
	log.Println("Perform login for user:", username)
	loginUrl, ppft, referer, cookies, err := getLoginPage(client)
	if err != nil {
		log.Println("Fail to login by error", err)
		return "", err
	}

	log.Println("Post login info to:", loginUrl)
	loginForm := buildLoginForm(username, password, ppft)
	loginReq, _ := http.NewRequest("POST", loginUrl, strings.NewReader(loginForm))
	loginReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("Referer", referer)
	for _, c := range cookies {
		loginReq.AddCookie(c)
	}
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Println("Fail to post login form")
		return "", err
	}
	defer loginResp.Body.Close()
	log.Println("Login Response:", loginResp.Status)
	postSrfDoc, err := goquery.NewDocumentFromReader(loginResp.Body)
	if err != nil {
		log.Println("Fail to parse response body HTML", err)
		return "", err
	}
	form := postSrfDoc.Find("form").First()
	oauthProxyUrl := form.AttrOr("action", "")
	oauthProxyForm := url.Values{}
	postSrfDoc.Find("form input").Each(func(i int, s *goquery.Selection) {
		oauthProxyForm.Add(s.AttrOr("name", ""), s.AttrOr("value", ""))
	})
	log.Println("Post Data to Oauth Proxy Url", oauthProxyUrl)
	oauthProxyReq, _ := http.NewRequest("POST", oauthProxyUrl, strings.NewReader(oauthProxyForm.Encode()))
	oauthProxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	oauthProxyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	oauthProxyReq.Header.Set("Referer", "https://login.live.com/")
	oauthProxyReq.Header.Set("Origin", "https://login.live.com")
	oauthProxyResp, err := client.Do(oauthProxyReq)
	if err != nil {
		log.Println("Fail to post to Oauth Proxy Url")
		return "", err
	}
	defer oauthProxyResp.Body.Close()
	oauthProxyDoc, err := goquery.NewDocumentFromReader(oauthProxyResp.Body)
	if err != nil {
		log.Fatal(err)
	}
	redirectUrl := oauthProxyDoc.Find("form").AttrOr("action", "")
	log.Println("Redirect Url:", redirectUrl)
	redirectForm := url.Values{}
	oauthProxyDoc.Find("form input").Each(func(i int, s *goquery.Selection) {
		redirectForm.Add(s.AttrOr("name", ""), s.AttrOr("value", ""))
	})
	redirectReq, _ := http.NewRequest("POST", redirectUrl, strings.NewReader(redirectForm.Encode()))
	redirectReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	redirectReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	redirectReq.Header.Set("Origin", "https://lw.skype.com")
	redirectReq.Header.Set("Referer", oauthProxyUrl)

	redirectResp, err := client.Do(redirectReq)
	if err != nil {
		log.Fatal(err)
	}
	defer redirectResp.Body.Close()
	log.Println("Redirect Response:", redirectResp.Status)

	finalDoc, err := goquery.NewDocumentFromReader(redirectResp.Body)
	if err != nil {
		log.Fatal(err)
	}

	skypeToken := ""
	finalDoc.Find("form input").Each(func(i int, s *goquery.Selection) {
		if s.AttrOr("name", "<no_name>") == "skypetoken" {
			skypeToken = s.AttrOr("value", "")
		}
	})
	return skypeToken, nil
}

func buildLoginForm(username string, password string, ppft string) string {
	loginForm := url.Values{}
	loginForm.Add("i13", "0")
	loginForm.Add("login", username)
	loginForm.Add("loginfmt", username)
	loginForm.Add("type", "11")
	loginForm.Add("LoginOptions", "3")
	loginForm.Add("lrt", "")
	loginForm.Add("ltrPartition", "")
	loginForm.Add("ltrPartition", "")
	loginForm.Add("hisRegion", "")
	loginForm.Add("hisScaleUnit", "")
	loginForm.Add("passwd", password)
	loginForm.Add("ps", "2")
	loginForm.Add("psRNGCDefaultType", "")
	loginForm.Add("psRNGCEntropy", "")
	loginForm.Add("psRNGCSLK", "")
	loginForm.Add("canary", "")
	loginForm.Add("ctx", "")
	loginForm.Add("hpgrequestid", "")
	loginForm.Add("PPFT", ppft)
	loginForm.Add("PPSX", "Passpo")
	loginForm.Add("NewUser", "1")
	loginForm.Add("FoundMSAs", "")
	loginForm.Add("fspost", "0")
	loginForm.Add("i21", "0")
	loginForm.Add("CookieDisclosure", "0")
	loginForm.Add("IsFidoSupported", "1")
	loginForm.Add("isSignupPost", "0")
	loginForm.Add("i2", "1")
	loginForm.Add("i17", "0")
	loginForm.Add("i18", "")
	loginForm.Add("i19", "1"+RandStringRunes(3))
	return loginForm.Encode()
}

func getLoginPage(client http.Client) (string, string, string, []*http.Cookie, error) {
	req, _ := http.NewRequest("GET", "https://web.skype.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Fail to parse login page body. Server return status:", resp.Status)
		return "", "", "", nil, err
	}

	body := string(bodyRaw)
	re := regexp.MustCompile("'(.*?)'")
	re1 := regexp.MustCompile("urlPost:'(.*?)'")
	loginUrl := strings.Trim(re.FindAllString(re1.FindAllString(body, -1)[0], -1)[0], "'")
	re2 := regexp.MustCompile("sFTTag:'(.*?)'")
	hiddenInput := strings.Trim(re.FindAllString(re2.FindAllString(body, -1)[0], -1)[0], "'")
	document, _ := goquery.NewDocumentFromReader(strings.NewReader(hiddenInput))
	ppft := document.Find("input").First().AttrOr("value", "")
	referer := resp.Request.URL.String()
	return loginUrl, ppft, referer, resp.Cookies(), nil
}

var letterRunes = []rune("0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
