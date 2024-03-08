package Server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (server *TVServer) GoogleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	googleConfig, err := SetUpConfig()
	if err != nil {
		fmt.Printf("Error setting up configuration: %v\n", err)
		return
	}
	url := googleConfig.AuthCodeURL("randomstate")
	http.Redirect(res, req, url, http.StatusSeeOther)
}

func SetUpConfig() (*oauth2.Config, error) {
	conf := &oauth2.Config{
		ClientID:     "KNMWUgBoWxm43xSHKRhfeF3eOOVJz6RH",
		ClientSecret: "cn-Im4B4kiqnmt0v70kuuEerbtQdeHkSr8Vw8fnJaXe16R4uqmDXyurwUFlLZrFD",
		RedirectURL:  "http://localhost:3000/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	if conf.ClientID == "" || conf.ClientSecret == "" || conf.RedirectURL == "" {
		return nil, fmt.Errorf("invalid configuration: ClientID, ClientSecret, and RedirectURL are required")
	}

	fmt.Println("conf ", conf)
	return conf, nil
}

func (server *TVServer) GoogleCallBack(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := req.URL.Query()["state"][0]
	if state != "randomstate" {
		fmt.Fprintln(res, "Invalid State Parameter!")
		return
	}

	code := req.URL.Query()["code"][0]
	googleConfig, _ := SetUpConfig()
	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Fprintln(res, "Error Occured: ", err)
	}
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Fprintln(res, "User data fetch failed")
	}
	userData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(res, "Json parsing failed.")
	}
	fmt.Fprintln(res, string(userData))
}
