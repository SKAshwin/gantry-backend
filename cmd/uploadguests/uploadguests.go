package main

//a simple command line program to upload a CSV of guest names, nric to the server
//given the server address

import (
	"bytes"
	"checkin"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	serverAddress := flag.String("addr", "http://localhost:3000", "The address of the server")
	eventID := flag.String("event", "", "The eventID of the event")
	filePath := flag.String("src", "guests.csv",
		"The CSV file with guest names, in a (nric,name) format, or (nric,name,tags) if -tags is supplied")
	token := flag.String("auth", "",
		"An authentication token with the rights to register new guests for the given event")
	username := flag.String("u", "", "Username for log in (needed if no authentication token)")
	loggerOutput := flag.String("out", "",
		"Where the output of the program will be dumped. Optional, if not specified output"+
			"dumped to standard output/the console")
	tags := flag.Bool("tags", false, "Use -tags if the CSV file is in a (nric,name,tags) format; tags should be comma separated, case insensitive. E.g. vip,confirmed will add the VIP and CONFIRMED tags to the guest in that row")
	versionNum := flag.Bool("v", false, "To get the version number of uploadguests")

	flag.Parse()

	if *versionNum {
		fmt.Println("uploadguests by MES Creators, Version 2.0. Compatible with Gantry by MES Version 1.3 and above.")
		return
	}
	if *eventID == "" {
		log.Fatal("Need to provide event ID (-event). See -h for help.")
	}
	if *token == "" && *username == "" {
		log.Fatal("Need authentication token or username. See -h for help")
	}
	if *loggerOutput != "" {
		out, err := os.Create(*loggerOutput)
		if err != nil {
			log.Fatal("Error opening file to write logs: " + err.Error())
		}
		log.SetOutput(out)
	}
	if *username != "" {
		fmt.Println("Attempting log in...")
		fmt.Print("Enter Password: ")
		bytePwd, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password: " + err.Error())
			return
		}
		pwd := strings.TrimSpace(string(bytePwd))
		reqBody := "{\"username\":\"" + *username + "\",\"password\":\"" + pwd + "\"}"
		//try the user log in, then the admin log in
		authURL := *serverAddress + "/api/v0/auth/users/login"
		resp, err := http.Post(authURL, "", strings.NewReader(reqBody))
		if err != nil {
			fmt.Println("Error fetching authentication: " + err.Error())
			return
		}
		//if failed, try admin
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Failed when trying to log in as user. Will log in as admin...")
			authURL = *serverAddress + "/api/v0/auth/admins/login"
			resp, err = http.Post(authURL, "", strings.NewReader(reqBody))
			if err != nil {
				fmt.Println("Error fetching authentication: " + err.Error())
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Log-in failed")
				return
			}
		}
		reply := struct {
			AccessToken string `json:"accessToken"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&reply)
		if err != nil {
			fmt.Println("Could not fetch auth token: " + err.Error())
			return
		}
		token = &reply.AccessToken
		fmt.Println("Authentication successful")
	}
	url := *serverAddress + "/api/v1-3/events/" + *eventID + "/guests"
	log.Println("Reading from " + *filePath + " and sending data to " + url)

	f, err := os.Open(*filePath)
	if err != nil {
		log.Fatal("Error opening CSV: " + err.Error())
		return
	}
	defer f.Close() // this needs to be after the err check
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV: " + err.Error())
	}

	tr := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 20,
	}
	client := &http.Client{Transport: tr}

	log.Println("Converting CSV into Guest Array JSON")
	guests := make([]checkin.Guest, len(lines))
	for i := 0; i < len(lines); i++ {
		guest := checkin.Guest{
			Name: lines[i][1],
			NRIC: lines[i][0],
		}
		if *tags {
			guest.Tags = extractTags(lines[i][2])
		}

		guests[i] = guest
	}
	guestsJSON, err := json.Marshal(guests)
	if err != nil {
		log.Fatal("Error marshalling CSV into guest array JSON : " + err.Error())
	}

	log.Println("Uploading guest array JSON to server")
	req, err := http.NewRequest("POST", url, bytes.NewReader(guestsJSON))
	if err != nil {
		log.Fatal("Error creating request to " + url + " :" + err.Error())
	}
	req.Header.Add("Authorization", "Bearer "+*token)

	resp, err := client.Do(req)
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

	log.Println("Response from server: " + reply.Message)
	log.Println("Finished uploading all guests")
}

func extractTags(tags string) []string {
	if tags == "" {
		return []string{}
	}
	tagArray := strings.Split(strings.ToUpper(tags), ",")
	for i, tag := range tagArray {
		tagArray[i] = strings.TrimSpace(tag)
	}

	return tagArray
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
