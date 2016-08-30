package main

import (
	"bytes"
	"io"
	"testing"
)

type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error { return nil }

func TestProcessWitResponseRestaurantSentiment(t *testing.T) {
	withJSON := stringToReadeClosser(restaurantSentiment)
	witResp := ProcessWitResponse(withJSON)

	intent := witResp.Entities.Intent
	if intent[0].Value != "restaurant:save" {
		t.Errorf("ProcessWitResponse didn't parse the 'intent' entities. Expected 'restaurant:save', we got: \n%+v\n", intent)
	}

	localSearchQry := witResp.Entities.LocalSearchQuery
	if localSearchQry[0].Value != "Sol y Luna" {
		t.Errorf("ProcessWitResponse didn't parse the 'localSearchQry' entities. Expected 'Sol y Luna', we got: \n%+v\n", localSearchQry)
	}

	sent := witResp.Entities.Sentiment
	if sent[0].Value != "negative" {
		t.Errorf("ProcessWitResponse didn't parse the 'sentiment' entities. Expected 'negative', we got: \n%+v\n", sent)
	}
}

func TestProcessWitResponsereminderAndLocation(t *testing.T) {
	withJSON := stringToReadeClosser(reminderAndLocation)
	witResp := ProcessWitResponse(withJSON)
	locations := witResp.Entities.Location
	if len(locations) != 1 {
		t.Errorf("ProcessWitResponse didn't parse the 'locations' entities. Expected just one entry in slice, we got: \n%+v\n", locations)
	}
	if locations[0].Type != "value" {
		t.Errorf("ProcessWitResponse didn't parse the 'locations' entities. Expected 'type' was 'value' , we got: \n%+v\n", locations[0].Value)
	}
	if locations[0].Confidence != 0.8164062572827963 {
		t.Errorf("ProcessWitResponse didn't parse the 'locations' entities. Expected 'confidence' was '0.8164062572827963' , we got: \n%+v\n", locations[0].Confidence)
	}
	if locations[0].Value != "grocery" {
		t.Errorf("ProcessWitResponse didn't parse the 'locations' entities. Expected 'value' was 'grocery' , we got: \n%+v\n", locations[0].Value)
	}

	reminders := witResp.Entities.Reminder

	if len(reminders) != 1 {
		t.Errorf("ProcessWitResponse didn't parse the 'reminders' entities. Expected just one entry in slice, we got: \n%+v\n", reminders)
	}
	if reminders[0].Type != "value" {
		t.Errorf("ProcessWitResponse didn't parse the 'reminders' entities. Expected 'type' was 'value' , we got: \n%+v\n", reminders[0].Value)
	}
	if reminders[0].Confidence != 0.9809287922999451 {
		t.Errorf("ProcessWitResponse didn't parse the 'reminders' entities. Expected 'confidence' was '0.9809287922999451' , we got: \n%+v\n", reminders[0].Confidence)
	}
	if reminders[0].Value != "get milk" {
		t.Errorf("ProcessWitResponse didn't parse the 'reminders' entities. Expected 'value' was 'get milk' , we got: \n%+v\n", reminders[0].Value)
	}
}

func stringToReadeClosser(s string) io.ReadCloser {
	return NopCloser{bytes.NewBufferString(s)}
}

func TestSanitizeQuerryStringStringLen(t *testing.T) {
	_, err := sanitizeQuerryString(string300)
	if err == nil {
		t.Error("FetchIntent did not return an error for a string input of 300 chars")
	}
	_, err = sanitizeQuerryString(string254)
	if err != nil {
		t.Errorf("FetchIntent returned an error for a string input of 254 chars %+v", err)
	}
}

const string300 = (`245485328217529591072968367825520430801937353549236235032205454278011159517553408301117871215897624083557692321819508308225339640853054008672033271569751783199322357002915818244872430853340789879400481978383988517251094914866992168126566388692301329752249123938027308855068750472072224632356977779896`)

const string254 = (`24548532821752959107296836782552043080193735354923623503220545427801115951755340830111787121589762408s3557692362408s355769232181950830822533964085305400867203327156975178319932235700291581824487243085334078987940048197838398851725109497222463235697777989`)

const reminderAndLocation = `{
  "msg_id" : "7efbe77d-13cd-4c76-9eec-dbe4081d9382",
  "_text" : "get milk from the grocery",
  "entities" : {
    "reminder" : [ {
      "confidence" : 0.9809287922999451,
      "type" : "value",
      "value" : "get milk",
      "suggested" : true
    } ],
    "location" : [ {
      "confidence" : 0.8164062572827963,
      "type" : "value",
      "value" : "grocery",
      "suggested" : true
    } ]
  }
}`

const restaurantSentiment = `{
  "msg_id" : "80d160f7-2122-4030-b22e-501afa1ca564",
  "_text" : "Sol y Luna is a bad restaurant for us",
  "entities" : {
    "local_search_query" : [ {
      "confidence" : 0.9784459645534703,
      "type" : "value",
      "value" : "Sol y Luna",
      "suggested" : true
    } ],
    "intent" : [ {
      "confidence" : 0.9995180263555402,
      "value" : "restaurant:save"
    } ],
    "sentiment" : [ {
      "confidence" : 0.9918119028290711,
      "value" : "negative"
    } ]
  }
}`
