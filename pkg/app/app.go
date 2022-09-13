package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	all_apps    = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?format=json"
	app_details = "https://store.steampowered.com/api/appdetails?appids="
	max_apps    = 6
)

type gameDetails struct {
	IsFree        bool   `json:"is_free"`
	HeaderImage   string `json:"header_image"`
	PriceOverview struct {
		PriceFormatted string `json:"final_formatted"`
	} `json:"price_overview"`
}

type Game struct {
	Appid  int    `json:"appid"`
	Name   string `json:"name"`
	Price  string `json:"price"`
	Image  string `json:"image"`
	IsFree bool   `json:"free"`
}

type gameList struct {
	Apps []Game `json:"apps"`
}

type games struct {
	Applist gameList `json:"applist"`
}

func GetAllGames() ([]Game, error) {
	r, err := http.Get(all_apps)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	apps := games{}
	err = json.Unmarshal(body, &apps)
	if err != nil {
		return nil, err
	}

	return apps.Applist.Apps, nil
}

func Format(apps []Game) []Game {
	var ret []Game
	for _, app := range apps {
		appName := cleanString(app.Name)
		if appName == "" || app.Appid == 0 {
			continue
		}

		ret = append(ret, Game{
			Name:  appName,
			Appid: app.Appid,
		})
	}

	return ret
}

func LookFor(key string, apps []Game) ([]Game, error) {
	var ret []Game

	if key == "" || len(key) <= 2 {
		return ret, nil
	}

	for _, app := range apps {
		if len(ret) == max_apps {
			return ret, nil
		}

		if strings.Contains(app.Name, key) {
			details, err := GetAppDetails(app.Appid)
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

func GetAllAppsDetails(apps []Game) ([]Game, error) {
	var ret []Game

	for _, app := range apps {
		details, err := GetAppDetails(app.Appid)
		if err != nil {
			return nil, err
		}

		app.Price = details.PriceOverview.PriceFormatted
		app.Image = details.HeaderImage
		app.IsFree = details.IsFree

		ret = append(ret, app)
	}

	return ret, nil
}

func GetAppDetails(appid int) (*gameDetails, error) {
	body, err := request(fmt.Sprintf("%s%d", app_details, appid))
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

	details := gameDetails{}
	err = json.Unmarshal(info, &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func GetAllAppsByName(key, tableName string, db *dynamodb.DynamoDB) ([]Game, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(tableName),
		FilterExpression: aws.String("contains(#name, :name)"),
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: aws.String(key),
			},
		},
	}

	result, err := db.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("could not get apps by [%s] : [%s]", key, err.Error())
	}

	games := new([]Game)

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &games)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal list of games: [%s]", err.Error())
	}

	return *games, nil
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

func cleanString(s string) string {
	nonAlphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(s, "")))), " ")
}
