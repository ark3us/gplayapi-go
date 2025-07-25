package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ark3us/gplayapi-go"
)

const sessionFile = "session.json"

func loadSession() (string, string, error) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return "", "", err
	}
	var sessionData struct {
		Email    string `json:"email"`
		AASToken string `json:"aas"`
	}
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return "", "", err
	}
	return sessionData.Email, sessionData.AASToken, nil
}

func main() {
	email, aasToken, err := loadSession()
	client, err := gplayapi.NewClientWithDeviceInfo(
		email,
		aasToken,
		gplayapi.Pixel8,
		"it",
		"it_IT",
	)
	if err != nil {
		log.Fatal(err)
	}
	app, err := client.GetAppDetails("com.discord")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("App details: %s (%d)\n", app.VersionName, app.VersionCode)
	deliveryData, err := client.Purchase(app.PackageName, app.VersionCode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(deliveryData.GetDownloadUrl())
	if split := deliveryData.SplitDeliveryData; split != nil {
		for _, s := range split {
			fmt.Printf("%s %s\n", s.GetName(), s.GetDownloadUrl())
		}
	}
}
