package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = os.Getenv("SPREADSHEET_ID")
var inList = false
var inTable = false
var first = true

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

func printFile(file string) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Could not read %s file: %s", file, err)
	} else {
		data := string(b)
		fmt.Println(data)
	}
}

func processRow(idx int, row []interface{}) {
	if inList {
		fmt.Println("</ul>")
		inList = false
	} else if inTable {
		fmt.Println("</table>")
		inTable = false
	}
	if isEqual(row[0], "x") {
		return
	}

	if !isEmpty(row[0]) {
		titleLevel, err := strconv.Atoi(row[0].(string))
		if err != nil {
			log.Printf("-- unexpected title level value at row %d, column 0 (found `%s`)\n", idx, row[0])
		}
		fmt.Println(title(titleLevel, row[1]))
		if len(row) >= 3 {
			fmt.Println(text(row[2]))
		}
		return
	}
	if len(row) < 3 {
		log.Printf("-- unexpected missing value at row %d, column 2\n", idx)
		return
	}
	if isEmpty(row[1]) {
		fmt.Println(text(row[2]))
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
	if level == 1 {
		if first == true {
			first = false
		} else {
			r += fmt.Sprintf("</div><br />")
		}
		r += fmt.Sprintf(`<div class="jt-Wrap--widthSpacer jt-Wrap--stylized">`)
	}
	r += fmt.Sprintf("<h%d>", level)
	r += rowValue.(string)
	r += fmt.Sprintf("</h%d>\n", level)
	return r
}

func tableLine(header bool, line []string) string {
		tags := [2]string{}
		if header {
			tags = [2]string{"<th>", "</th>"}
		} else {
			tags = [2]string{"<td>", "</td>"}
		}
		r := "<tr>\n"
		for i := range line {
			r += fmt.Sprintf("%s%s%s\n", tags[0], line[i], tags[1])
		}
		r += "</tr>\n"
		return r
}

func text(rowValue interface{}) string {
	r := ""
	row := rowValue.(string)
	lines := strings.Split(row, "\n")
	for i := range lines {
		if lines[i] != "" {
			if strings.Contains(lines[i], "**") {
				s := ""
				a := strings.Split(lines[i], "**")
				s += text(a[0])
				s += fmt.Sprintf("<strong>%s</strong>\n", a[1])
				s += text(strings.Join(a[2:], ""))
				lines[i] = s
			}
			if strings.Contains(lines[i], " | ") {
				s := ""
				l := strings.Split(lines[i], " | ")
				if !inTable {
					s += "<table>\n"
					s += tableLine(true, l)
					inTable = true
				} else {
					s += tableLine(false, l)
				}
				lines[i] = s
			} else if inTable {
				r += fmt.Sprintf("</table>\n")
				inTable = false
			}
			if strings.HasPrefix(lines[i], "-") {
				if !inList {
					r += fmt.Sprintf("<ul>\n")
					inList = true
				}
				r += fmt.Sprintf("<li>%s</li>\n", strings.TrimLeft(lines[i], "- "))
			} else {
				if inList {
					r += fmt.Sprintf("</ul>\n")
					inList = false
				}
				if strings.HasPrefix(lines[i], "#") {
					titleLevel := strings.LastIndex(lines[i], "#")
					r += title(titleLevel, strings.TrimLeft(lines[i], "# "))
				} else {
					r += fmt.Sprintf("\n<p>%s</p>\n", lines[i])
				}
			}
		}
	}
	return r
}

func isEmpty(rowValue interface{}) bool {
	return len(rowValue.(string)) == 0
}

func isEqual(rowValue interface{}, s string) bool {
	return rowValue.(string) == s
}
