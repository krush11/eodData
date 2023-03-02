package utils

import "net/http"

var BaseURL string = "https://www.nseindia.com"

func ReqConfig() *http.Request {
	req, _ := http.NewRequest("GET", BaseURL, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", "www.nseindia.com")
	req.Header.Add("Referer", "https://www.nseindia.com/get-quotes/equity")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")
	req.Header.Add("X-Forwarded-For", "2401:4900:47fa:df96:5efb:1767:23c6:e1b1")
	
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}
	
	cookies := req.Cookies()
	for i := 0; i < len(cookies); i++ {
		for j := i + 1; j < len(cookies); j++ {
			if cookies[i].Name == cookies[j].Name {
				cookies = append(cookies[:j], cookies[j+1:]...)
				j--
			}
		}
	}
	req.Header.Del("Cookie")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	
	return req
}
