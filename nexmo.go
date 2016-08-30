package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

//NexmoHandler handles the GET requests from nexmo
func NexmoHandler(w http.ResponseWriter, r *http.Request) {
	//A sample request from the nexmo service is
	//?msisdn=19150000001&to=12108054321
	//&messageId=000000FFFB0356D1&text=This+is+an+inbound+message
	//&type=text&message-timestamp=2012-08-19+20%3A38%3A23&concat=true //if long msg, could be missing
	//concat-ref	The transaction reference. All parts of this message share this concat-ref.	If concat is True
	//concat-total	The number of parts in this concatenated message.	If concat is True
	//concat-part	The number of this part in the message. Counting starts at 1.	If concat is True
	//So we read all those parameters
	userCell := r.FormValue("msisdn")
	messageID := r.FormValue("messageId")
	text := r.FormValue("text")
	typ := r.FormValue("type")
	timestamp := r.FormValue("message-timestamp")
	go func() {
		if len(text) > 0 && typ == "text" {
			intent, err := FetchIntent(text)
			if err != nil {
				log.Printf("Error: %+v", err)
			} else {
				ret := ProcessIntent(intent)
				log.Printf("We got messageID: %v on %v ", messageID, timestamp)
				log.Printf("Wit gave us: %+v ", ret)
				log.Printf("Intent: %+v ", ret.Intent)
				log.Printf("LocalSearchQuery: %+v ", ret.LocalSearchQuery)
				log.Printf("Sentiment: %+v ", ret.Sentiment)
				log.Printf("Food Or Drink: %+v ", ret.FoodOrDrink)
				payload := &RestaurantSentimentKey{
					UserPhone: strings.ToLower(userCell),
					Name:      strings.ToLower(ret.LocalSearchQuery.Value),
					Location:  "all",
					Food:      strings.ToLower(ret.FoodOrDrink.Value),
					WitMsgID:  intent.MsgID,
					FullMsg:   text,
				}
				if ret.Intent.Value == "restaurant:save" && payload.Name != "" && ret.Sentiment.Value != "" {
					log.Println("Saving to db...")
					if payload.Food == "" {
						payload.Food = "all"
					}
					err = payload.save(db, strings.ToLower(ret.Sentiment.Value))
					if err != nil {
						log.Printf("error saving %+v to csv, got: %s\n", payload, err)
					}
					sendSMS(userCell, "Got it.")
				} else {
					err = payload.save(db+".error", strings.ToLower(ret.Sentiment.Value))
					if err != nil {
						log.Printf("error saving %+v to csv.error, got: %s\n", payload, err)
					}
					sendSMS(userCell, "Hm, will need to reread this one.")
				}
			}

		} else {
			log.Print("Error: we got a blank text message")
		}
	}()
	w.WriteHeader(http.StatusOK)
}

func sendSMS(to, msg string) error {

	baseURL, err := url.Parse("https://rest.nexmo.com/sms/json?")
	if err != nil {
		log.Println("error while parsing url: ", err)
	}

	params := url.Values{}
	params.Add("api_key", config.NexmoAPIKey)
	params.Add("api_secret", config.NexmoAPISecret)
	params.Add("to", to)
	params.Add("from", config.NexmoFromNumber)
	params.Add("text", msg)
	baseURL.RawQuery = params.Encode()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL.String(), bytes.NewBufferString(params.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params.Encode())))
	res, err := client.Do(req)

	if err != nil {
		log.Fatalf("Sending sms using nexmo gave: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("Something went really wrong sending sms, got: ", err)
		return nil
	}
	var jsonResponse NexmoResponse
	b, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(b, &jsonResponse)
	if err != nil {
		log.Println("error parsing json: ", err)
	}
	for _, v := range jsonResponse.Messages {
		if v.Status == "0" {
			log.Printf("all went well, response was: %+v\n", v)
		}
		if v.Status != "0" {
			log.Printf("there was a problem, response was: %+v\n", v)
			if v.Status == "1" {
				log.Println("waiting 1 second and retrying  ...")
				time.Sleep(1 * time.Second)
				sendSMS(to, msg)
			}
			return errors.New(v.Status)
		}
	}

	return nil
}

//NexmoResponse represents the response we get from nexmo after sending an sms
type NexmoResponse struct {
	MessageCnt string `json:"message-count"`
	Messages   []struct {
		Status           string
		MessageID        string `json:"message-id"`
		To               string
		ClientRef        string `json:"client-ref"`
		RemainingBalance string `json:"remaining-balance"`
		MessagePrice     string `json:"message-price"`
		Network          string `json:"network"`
		ErrorText        string `json:"error-text"`
	}
}

//RestaurantSentimentKey is used to save data into BoltDB
type RestaurantSentimentKey struct {
	UserPhone string
	Name      string
	Location  string
	Food      string
	WitMsgID  string
	FullMsg   string
}

func (key *RestaurantSentimentKey) save(filePath string, sentiment string) error {

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("error opening file %s, got: %s", filePath, err)
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)

	record := []string{time.Now().String(), key.UserPhone, key.Name, key.Food, key.Location, key.WitMsgID, sentiment, key.FullMsg}
	if err := w.Write(record); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

//http://192.168.1.10:1212/sms?msisdn=18283930032&to=18285728494&messageId=000000FFFB0356D1&text=the+chicken+at+El+Paso+is+great&type=text&message-timestamp=2012-08-19+20%3A38%3A23
//http://127.0.0.1:1212/sms?msisdn=18283930032&to=18285728494&messageId=000000FFFB0356D1&text=the+chicken+at+El+Paso+is+great&type=text&message-timestamp=2012-08-19+20%3A38%3A23
