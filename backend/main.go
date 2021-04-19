package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"time"
)

func main() {
	http.ListenAndServe(":8000", ServeService())
}

type ErrorMessage struct {
	Error string `json:"error"`
}

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type Plant struct {
	Link   []string `json:"url"`
	Cookie string   `json:"cookie"`
}

type Success struct {
	Key   string `json:"key"`
	Count int    `json:"success"`
	Limit int    `json:"limit"`
	Fail  int    `json:"fail"`
}

func WaterAnon(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == "OPTIONS" {
		return
	}
	var plants Plant
	err := json.NewDecoder(r.Body).Decode(&plants)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorMessage{Error: "Invalid JSON"})
		return
	}
	cookie := plants.Cookie

	if cookie == "" {
		cookie = "csrftoken=ryDrEZbsfa7s8KTlKtLUctWV0HHWszsL; SPC_F=DddGaoOzqAHsVHMUCfh7VVIhLBaKg2K6; share_key_31901183_2020-05-02=1"
	}

	output := make([]Success, 0)

	for _, plantLink := range plants.Link {
		u, _ := url.Parse(plantLink)
		values, _ := url.ParseQuery(u.RawQuery)

		plantKey := values.Get("skey")
		plantChannel := values.Get("schannel")
		
		successCount := 0
		limitCount := 0
		failCount := 0

		if plantKey == "" {
			plantKey = plantLink
			failCount = 1
		} else {
			for i := 1; i <= 5; i++ {
				if i != 1 {
					time.Sleep(500 * time.Millisecond)
				}
				msg := WaterAnonCall(plantKey, plantChannel, cookie)
				if msg == "success" {
					successCount++
				} else if msg == "accept anonymous user help count limited" {
					limitCount++
					break
				} else {
					failCount++
					break
				}
			}
		}
		var succ Success = Success{Key: plantKey, Count: successCount, Limit: limitCount, Fail: failCount}
		output = append(output, succ)
	}
	json.NewEncoder(w).Encode(output)
	return
}

func WaterAnonCall(key string, channel string, cookie string) string {
	url := "https://games.shopee.sg/farm/api/friend/anonymous/help"
	requestBody, err := json.Marshal(map[string]string{
		"channel":  channel,
		"shareKey": key,
	})

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	request.Header.Set("accept", "application/json, text/plain, */*")
	request.Header.Set("authority", "games.shopee.sg")
	request.Header.Set("pragma", "no-cache")
	request.Header.Set("cache-control", "no-cache")
	request.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.129 Safari/537.36")
	request.Header.Set("content-type", "application/json;charset=UTF-8")
	request.Header.Set("origin", "https://games.shopee.sg")
	request.Header.Set("sec-fetch-site", "same-origin")
	request.Header.Set("sec-fetch-mode", "cors")
	request.Header.Set("sec-fetch-dest", "empty")
	request.Header.Set("referer", url)
	request.Header.Set("accept-language", "en-US,en;q=0.9")
	request.Header.Set("cookie", cookie)

	if err != nil {
		return "failed"
	}

	client := http.Client{
		Timeout: time.Duration(4 * time.Second),
	}
	resp, err := client.Do(request)

	defer resp.Body.Close()
	var body Resp
	err = json.NewDecoder(resp.Body).Decode(&body)
	return body.Msg
}

func ServeService() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/water/anon", WaterAnon).Methods("POST", "OPTIONS")
	return router
}
