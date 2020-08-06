package worker

import "net/http"

func (w *PollingWorker) setRequestHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "skypetoken="+w.skypeToken)
	req.Header.Set("RegistrationToken", w.registrationToken)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")
	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=VN; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
}
func (w *PollingWorker) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
	req.Header.Set("Referer", "https://web.skype.com/")
	req.Header.Set("Origin", "https://web.skype.com")
	req.Header.Set("BehaviorOverride", "redirectAs404")
	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=Unknown; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
}
