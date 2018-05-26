package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = os.Getenv("SPREADSHEET_ID")

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	readRange := os.Getenv("SPREADSHEET_RANGE")
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for i, row := range resp.Values {
			processRow(i, row)
		}
	}
}

func processRow(idx int, row []interface{}) {
	if isEqual(row[0], "x") {
		return
	}
	if !isEmpty(row[0]) {
		titleLevel, err := strconv.Atoi(row[0].(string))
		if err != nil {
			fmt.Printf("-- unexpected title level value at row %d, column 0 (found `%s`)\n", idx, row[0])
		}
		fmt.Println(title(titleLevel, row[1]))
		return
	}
	if isEmpty(row[1]) {
		fmt.Println(text(row[2]))
		return
	}
	if len(row) < 3 {
		fmt.Printf("-- unexpected missing value at row %d, column 2\n", idx)
		return
	}
	if isEqual(row[2], "N/A") {
		return
	}
	fmt.Println(title(4, row[1]))
	fmt.Println(text(row[2]))
}

func title(level int, rowValue interface{}) string {
	r := "\n"
	for i := 0; i < level; i++ {
		r += "#"
	}
	r += " " + rowValue.(string) + "\n"
	return r
}

func text(rowValue interface{}) string {
	return "\n" + rowValue.(string) + "\n"
}

func isEmpty(rowValue interface{}) bool {
	return len(rowValue.(string)) == 0
}

func isEqual(rowValue interface{}, s string) bool {
	return rowValue.(string) == s
}
