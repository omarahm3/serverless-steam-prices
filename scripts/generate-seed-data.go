package main

import (
	"encoding/json"
	"os"

	"github.com/omarahm3/serverless-steam-prices/pkg/app"
)

const DATA_PATH = "./seed/apps.json"

func main() {
	apps, err := app.GetAllGames()
	check(err)
	apps = app.Format(apps)
	err = save(apps)
	check(err)
}

func save(apps []app.Game) error {
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

func check(err error) {
	if err == nil {
		return
	}

	panic(err)
}
