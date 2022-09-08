package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	ALL_APPS    = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?format=json"
	APP_DETAILS = "https://store.steampowered.com/api/appdetails?appids="
	MAX_APPS    = 5
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

type Json map[string]interface{}

type AppDetails struct {
	IsFree        bool   `json:"is_free"`
	HeaderImage   string `json:"header_image"`
	PriceOverview struct {
		PriceFormatted string `json:"final_formatted"`
	} `json:"price_overview"`
}

type App struct {
	Appid  int    `json:"appid"`
	Name   string `json:"name"`
	Price  string `json:"price"`
	Image  string `json:"image"`
	IsFree bool   `json:"free"`
}

type AppList struct {
	Apps []App `json:"apps"`
}

type Apps struct {
	Applist AppList `json:"applist"`
}

type (
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
)

// Handler function Using AWS Lambda Proxy Request
func Handler(r Request) (Response, error) {
	// Get the path parameter that was sent
	query := r.QueryStringParameters["query"]

	if query == "" {
		return response(Json{"message": "you must supply 'query' parameter"}, 400)
	}

	if len(query) <= 2 {
		return response(Json{"message": "query must be more than 2 characters"}, 404)
	}

	apps, err := getAllApps()
	if err != nil {
		return response(Json{"message": "error occurred while retrieving steam apps"}, 400)
	}

	apps = format(apps)

	found, err := lookFor(query, apps)
	if err != nil {
		return response(Json{"message": "error occurred while getting app details"}, 400)
	}

	return response(Json{
		"total": len(found),
		"apps":  found,
	}, 200)
}

func main() {
	lambda.Start(Handler)
}

func lookFor(key string, apps []App) ([]App, error) {
	var ret []App

	if key == "" || len(key) <= 2 {
		return ret, nil
	}

	for _, app := range apps {
		if len(ret) > MAX_APPS {
			return ret, nil
		}

		if strings.Contains(app.Name, key) {
			details, err := getAppDetails(app.Appid)
			if err != nil {
				return nil, err
			}

			app.Price = details.PriceOverview.PriceFormatted
			app.Image = details.HeaderImage
			app.IsFree = details.IsFree

			ret = append(ret, app)
		}
	}

	return ret, nil
}

func format(apps []App) []App {
	ret := make([]App, len(apps))
	for _, app := range apps {
		appName := cleanString(app.Name)
		if appName == "" {
			continue
		}

		ret = append(ret, App{
			Name:  appName,
			Appid: app.Appid,
		})
	}

	return ret
}

// TODO apply a cache over this function for sure
func getAppDetails(appid int) (*AppDetails, error) {
	body, err := request(fmt.Sprintf("%s%d", APP_DETAILS, appid))
	if err != nil {
		return nil, err
	}

	var data map[string]map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	info, err := json.Marshal(data[strconv.Itoa(appid)]["data"])
	if err != nil {
		return nil, err
	}

	details := AppDetails{}
	err = json.Unmarshal(info, &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func getAllApps() ([]App, error) {
	r, err := http.Get(ALL_APPS)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	apps := Apps{}
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return nil, err
	}

	return apps.Applist.Apps, nil
}

func cleanString(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(s, "")))), " ")
}

func request(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func response(body Json, status int) (Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return Response{StatusCode: 400}, err
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, b)

	return Response{
		StatusCode: status,
		Body:       buf.String(),
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "GET",
		},
	}, nil
}
