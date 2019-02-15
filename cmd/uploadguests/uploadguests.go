package main

//a simple command line program to upload a CSV of guest names, nric to the server
//given the server address

import (
	"bytes"
	"checkin"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 5 {
		log.Println("uploadguests [serverAddress] [eventID] [filePathToCSV] [authToken]")
		log.Println("Example:")
		log.Println("uploadguests http://checkin.com 85fdbd2a-e879-4c1d-b4ae-f70dacbc7816 Documents/guests.csv")
		log.Println("Do not include a / after the server address")
		log.Println("guests.csv should be formatted as:")
		log.Println("[last5DigitsOfNRIC],[Name]")
		log.Println("Name displayed to guest upon check in")
		log.Println("Do not include a title row in the CSV")
	}
	serverAddress := os.Args[1]
	eventID := os.Args[2]
	filePath := os.Args[3]
	token := os.Args[4]

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening CSV: " + err.Error())
		return
	}
	defer f.Close() // this needs to be after the err check

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV: " + err.Error())
	}
	for i := 0; i < len(lines); i++ {
		guest := checkin.Guest{
			Name: lines[i][1],
			NRIC: lines[i][0],
		}
		guestJSON, err := json.Marshal(guest)
		if err != nil {
			log.Fatal("Error marshalling CSV into JSON for " + guest.NRIC + ", " + guest.Name + " : " +
				err.Error())
		}
		url := serverAddress + "/api/v0/events/" + eventID + "/guests"
		log.Println("POST (" + guest.NRIC + ", " + guest.Name + ") to " + url)

		req, err := http.NewRequest("POST", url, bytes.NewReader(guestJSON))
		if err != nil {
			log.Fatal("Error creating request to " + url + " :" + err.Error())
		}
		req.Header.Add("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal("Error posting to " + url + " :" + err.Error())
		}
		defer resp.Body.Close()

		reply := struct {
			Message string `json:"message"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&reply)
		if err != nil {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				log.Fatal("wtf: " + err2.Error())
			}
			log.Println(string(body))
			log.Fatal("Error reading response: " + err.Error())
		}

		log.Println("Response to registering (" + guest.NRIC + ", " + guest.Name + "):" + reply.Message)
	}
}
