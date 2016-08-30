package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// WitVersion is the wit api version we support
const WitVersion = "20160824"

//FetchIntent is the whole go wit wrapper, if you call it that.
//We send the query string to wit, parse the result json
//into a struct and return it.
func FetchIntent(str string) (WitMessage, error) {

	str, err := sanitizeQuerryString(str)
	if err != nil {
		return WitMessage{}, err
	}

	url := fmt.Sprintf("https://api.wit.ai/message?v=%s&q=%s", WitVersion, str)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.WitAccessToken))
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Requesting wit's api gave: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("Something went really wrong with the response from Wit.ai")
		errMsg := "Sorry, the machine learning service I use for my brain went down, @Diego: check the logs, there may be something for you there."
		return WitMessage{}, errors.New(errMsg)
	}
	return ProcessWitResponse(res.Body), nil
}

func sanitizeQuerryString(str string) (string, error) {
	if len(url.QueryEscape(str)) > 255 {
		errMsg := "Sorry, I can only read up to 256 characters and I didn't want to just ignore the end of your message."
		return "", errors.New(errMsg)
	}
	return url.QueryEscape(str), nil
}

//ProcessWitResponse gets the raw response from the http request, and
//returns a WitMessage with all the information we got from Wit
func ProcessWitResponse(message io.ReadCloser) WitMessage {
	intent, _ := ioutil.ReadAll(message)
	var jsonResponse WitMessage
	err := json.Unmarshal(intent, &jsonResponse)
	if err != nil {
		log.Println("error parsing json: ", err)
	}
	log.Printf("plain text json was: \n\n%+v", string(intent[:]))

	return jsonResponse

}

//ProcessIntent gets the json parsed result from wit.ai and
//depending on the intent, it calls the right service.
func ProcessIntent(jsonResponse WitMessage) WitResponse {
	var localSearchQry WitSingleResponse
	var intent WitSingleResponse
	var sentiment WitSingleResponse
	var foodOrDrink WitSingleResponse

FoodOrDrinkLoop:
	for _, row := range jsonResponse.Entities.FoodOrDrink {
		if row.Confidence > 0.80 {
			foodOrDrink = row
			break FoodOrDrinkLoop
		}
	}

LocalSearchQryLoop:
	for _, row := range jsonResponse.Entities.LocalSearchQuery {
		if row.Confidence > 0.80 {
			localSearchQry = row
			break LocalSearchQryLoop
		}
	}

IntentLoop:
	for _, row := range jsonResponse.Entities.Intent {
		if row.Confidence > 0.80 {
			intent = row
			break IntentLoop
		}
	}

SentimentLoop:
	for _, row := range jsonResponse.Entities.Sentiment {
		if row.Confidence > 0.80 {
			sentiment = row
			break SentimentLoop
		}
	}

	return WitResponse{
		LocalSearchQuery: localSearchQry,
		FoodOrDrink:      foodOrDrink,
		Intent:           intent,
		Sentiment:        sentiment,
	}
}

//These make up the different parts of the wit result
//There are more options, but I'm using only these so far.

//WitMessage represents the payload we get from Wit as a response to processing
//the text or voice file we sent.
type WitMessage struct {
	MsgID    string `json:"msg_id"`
	MsgBody  string `json:"_text"`
	Entities WitMessageEntities
}

//WitMessageEntities contains all the possible entities we process from Wit
type WitMessageEntities struct {
	Location         []WitSingleResponse `json:",omitempty"`
	Reminder         []WitSingleResponse `json:",omitempty"`
	LocalSearchQuery []WitSingleResponse `json:"local_search_query,omitempty"`
	FoodOrDrink      []WitSingleResponse `json:"food_or_drink,omitempty"`
	Sentiment        []WitSingleResponse `json:",omitempty"`
	Intent           []WitSingleResponse `json:",omitempty"`
	Datetime         []WitDatetime       `json:",omitempty"`
}

//WitSingleResponse is the Location entity
type WitSingleResponse struct {
	Confidence float64
	Type       string
	Value      string
	Suggested  bool
}

//WitDatetime is the Location entity
type WitDatetime struct {
	Confidence float64
	Type       string
	Value      string
	Grain      string
	From       WitDateTimeFromTo
	To         WitDateTimeFromTo
}

//WitDateTimeFromTo is the on_off entity
type WitDateTimeFromTo struct {
	Value string
	Grain string
}

//WitNumber is the wit/number entity
type WitNumber struct {
	End   int
	Start int
	Value int
	Body  string
}

//WitResponse holds just the information you need to act on each intent
type WitResponse struct {
	Location         WitSingleResponse `json:",omitempty"`
	Reminder         WitSingleResponse `json:",omitempty"`
	LocalSearchQuery WitSingleResponse `json:"local_search_query,omitempty"`
	FoodOrDrink      WitSingleResponse `json:"food_or_drink,omitempty"`
	Sentiment        WitSingleResponse `json:",omitempty"`
	Intent           WitSingleResponse `json:",omitempty"`
	Datetime         WitDatetime       `json:",omitempty"`
	Error            witError
}

type witError struct {
	msg string
}
