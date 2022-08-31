package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type vpnUser struct {
	Email    string `json:"email"`
	Disabled bool   `json:"disabled"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

func main() {
	url := "https://your.http.endpoint.com"

	vpnUsers := []vpnUser{}

	ctx := context.Background()

	// Step 1 : Update the audience with your IAP CLIENT ID.
	// Looks like this : IAP_CLIENT_ID.apps.googleusercontent.com
	audience := "IAP_CLIENT_ID.apps.googleusercontent.com"

	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		panic(err)
	}

	method := "GET"
	orgId := "xxxxxxxx"
	path := "/user/" + orgId

	// Step 2 : Obviously, update your api token and secret
	api_token := ""
	api_secret := ""
	now := time.Now()
	auth_timestamp := strconv.FormatInt(now.Unix(), 10)
	uuidString := uuid.New().String()
	// Important : Pritunl API does not like if the uuid string has dashes in them
	auth_nonce := strings.Replace(uuidString, "-", "", -1)

	auth_string_items := []string{api_token, auth_timestamp, auth_nonce, strings.ToUpper(method), path}
	auth_string := strings.Join(auth_string_items[:], "&")

	h := hmac.New(sha256.New, []byte(api_secret))
	h.Write([]byte(auth_string))

	auth_signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	request, err := http.NewRequest("GET", url+path, nil)

	request.Header.Set("Auth-Token", api_token)
	request.Header.Set("Auth-Timestamp", auth_timestamp)
	request.Header.Set("Auth-Nonce", auth_nonce)
	request.Header.Set("Auth-Signature", auth_signature)

	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	fmt.Println(response.StatusCode)

	err = json.NewDecoder(response.Body).Decode(&vpnUsers)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range vpnUsers {
		if u.Type == "client" {
			log.Printf("%s -> %s \n", u.Email, u.Name)
		}

	}
}
