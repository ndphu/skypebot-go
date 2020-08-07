package utils

//
//func PrepareMediaRequestHeaders(req *http.Request, transactionId string) {
//	req.Header.Set("TransactionId", transactionId)
//	req.Header.Set("Authorization", "skype_token "+config.Get().SkypeToken())
//	req.Header.Set("X-Client-Version", "1418/8.62.0.83//")
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
//	req.Header.Set("Referer", "https://web.skype.com/")
//	req.Header.Set("Origin", "https://web.skype.com")
//}
//
//func SetRequestHeaders(req *http.Request) {
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authentication", "skypetoken="+config.Get().SkypeToken())
//	req.Header.Set("RegistrationToken", config.Get().RegistrationToken())
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
//	req.Header.Set("Referer", "https://web.skype.com/")
//	req.Header.Set("Origin", "https://web.skype.com")
//	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=VN; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
//}
//
//func SetEndpointHeader(req *http.Request) {
//	req.Header.Set("EndpointId", config.Get().CurrentEndpoint())
//}
//
//func SetDefaultHeaders(req *http.Request) {
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36")
//	req.Header.Set("Referer", "https://web.skype.com/")
//	req.Header.Set("Origin", "https://web.skype.com")
//	req.Header.Set("BehaviorOverride", "redirectAs404")
//	req.Header.Set("ClientInfo", "os=Windows; osVer=10; proc=x86; lcid=en-US; deviceType=1; country=Unknown; clientName=skype4life; clientVer=1418/8.62.0.83//skype4life; timezone=Asia/Bangkok")
//}
