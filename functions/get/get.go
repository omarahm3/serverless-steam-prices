package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const ALL_APPS = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=STEAMKEY&format=json"

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

type Json map[string]interface{}

type App struct {
	Appid int    `json:"appid"`
	Name  string `json:"name"`
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
	found := lookFor(query, apps)

	return response(Json{
		"total": len(found),
		"apps":  found,
	}, 200)
}

func main() {
	lambda.Start(Handler)
}

func lookFor(key string, apps []App) []App {
	var ret []App

	if key == "" || len(key) <= 2 {
		return ret
	}

	for _, app := range apps {
		if strings.Contains(app.Name, key) {
			ret = append(ret, app)
		}
	}

	return ret
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
