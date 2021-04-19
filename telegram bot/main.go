package main

import (
	"github.com/joho/godotenv"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
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

type ErrorMessage struct {
	Error string `json:"error"`
}
// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

  // load .env file
  err := godotenv.Load(".env")

  if err != nil {
    log.Fatalf("Error loading .env file")
  }

  return os.Getenv(key)
}

func WaterAnon(plants []string) []Success {

	cookie := "csrftoken=ryDrEZbsfa7s8KTlKtLUctWV0HHWszsL; SPC_F=DddGaoOzqAHsVHMUCfh7VVIhLBaKg2K6; share_key_31901183_2020-05-02=1"

	output := make([]Success, 0)

	for _, plantLink := range plants {
		// CB
		// Added URL split here to read reat /water http://url.com
		newURl := strings.Split(plantLink, "/water ")
		u, _ := url.Parse(newURl[1])
		values, _ := url.ParseQuery(u.RawQuery)

		plantKey := values.Get("skey")
		plantChannel := values.Get("schannel")

		successCount := 0
		limitCount := 0
		failCount := 0
		for i := 1; i <= 5; i++ {
			if i != 1 {
				time.Sleep(500 * time.Millisecond)
			}
			msg := WaterAnonCall(plantKey, plantChannel, cookie)
			if msg == "success" {
				successCount++
			} else if msg == "accept anonymous user help count limited" {
				limitCount++
			} else {
				failCount++
				break
			}
		}
		var succ Success
		succ = Success{Key: plantKey, Count: successCount, Limit: limitCount, Fail: failCount}
		output = append(output, succ)
	}
	return output
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

// This handler is called everytime telegram sends us a webhook event
func Handler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	// Check if the message contains the word "marco"
	// if not, return without doing anything
	// if !strings.Contains(strings.ToLower(body.Message.Text), "marco") {
	// 	return
	// }
	output := make([]Success, 0)
	fmt.Println(strings.Split(body.Message.Text, "\n"))

	// CB
	// Added If statement to see such command exists inside the message
	if strings.Contains(body.Message.Text, "/water") {
		output = WaterAnon(strings.Split(body.Message.Text, "\n"))
	}

	// If the text contains marco, call the `sayPolo` function, which
	// is defined below
	if err := sayPolo(body.Message.Chat.ID, output); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("reply sent")
}

//The below code deals with the process of sending a response message
// to the user

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

// sayPolo takes a chatID and sends "polo" to them
func sayPolo(chatID int64, output []Success) error {
	// Create the request body struct
	beautifiedOutput := ""
	for _, s := range output {
		beautifiedOutput += "Key: " + s.Key
		beautifiedOutput += "\nSuccess: " + strconv.Itoa(s.Count)
		beautifiedOutput += "\nLimit: " + strconv.Itoa(s.Limit)
		beautifiedOutput += "\nFail: " + strconv.Itoa(s.Fail) + "\n\n"
	}
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   beautifiedOutput,
	}
	// Create the JSON body from the struct
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	keyenv := goDotEnvVariable("TELEGRAM_KEY")
	// Send a post request with your token
	res, err := http.Post("https://api.telegram.org/bot" + keyenv + "/sendMessage", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}

	return nil
}

// FInally, the main funtion starts our server on port 3000
func main() {
	http.ListenAndServe(":3000", http.HandlerFunc(Handler))
}
