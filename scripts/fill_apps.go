package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	ALL_APPS  = "http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=STEAMKEY&format=json"
	DATA_PATH = "./store.json"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

type JSONApp map[string]int

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

func main() {
	apps, err := getAllApps()
	check(err)
	apps = format(apps)
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

func check(err error) {
	if err == nil {
		return
	}

	panic(err)
}
