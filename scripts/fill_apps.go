package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	ALL_APPS    = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=STEAMKEY&format=json"
	APP_DETAILS = "https://store.steampowered.com/api/appdetails?appids="
	DATA_PATH   = "./store.json"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

type JSONApp map[string]int

type AppDetails struct {
	IsFree        bool   `json:"is_free"`
	HeaderImage   string `json:"header_image"`
	PriceOverview struct {
		PriceFormatted string `json:"final_formatted"`
	} `json:"price_overview"`
}

type App struct {
	Appid int    `json:"appid"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Image string `json:"image"`
}

type AppList struct {
	Apps []App `json:"apps"`
}

type Apps struct {
	Applist AppList `json:"applist"`
}

func main() {
	apps, err := getAllApps()
	check(err)
	apps, err = format(apps)
	check(err)
	err = save(apps)
	check(err)
}

func save(apps []App) error {
	f, err := os.OpenFile(DATA_PATH, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	str, err := json.MarshalIndent(apps, "", "\t")
	if err != nil {
		return err
	}

	_, err = f.Write(str)
	if err != nil {
		return err
	}

	return nil
}

func format(apps []App) ([]App, error) {
	ret := make([]App, len(apps))
	for _, app := range apps {
		appName := cleanString(app.Name)
		if appName == "" {
			continue
		}

		details, err := getAppDetails(app.Appid)
		if err != nil {
			return nil, err
		}

		ret = append(ret, App{
			Name:  appName,
			Appid: app.Appid,
			Price: details.PriceOverview.PriceFormatted,
			Image: details.HeaderImage,
		})
	}

	return ret, nil
}

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
	body, err := request(ALL_APPS)
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
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(s, "")))), " ")
}

func check(err error) {
	if err == nil {
		return
	}

	panic(err)
}
