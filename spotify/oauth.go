package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

type OAuth struct {
	me      string
	token   string
	refresh string
}

func (c *Client) Authenticate() (err error) {
	c.oauth, err = c.getOAuth()
	if err != nil {
		return Error{"(*Client).Authenticate", "(*Client).getOAuth", err}
	}

	c.oauth.me, err = c.getUserID()
	if err != nil {
		return Error{"(*Client).Authenticate", "(*Client).getUserID", err}
	}

	return nil
}

func (c *Client) getToken() (string, error) {
	req, err := http.NewRequest("POST", ACC_URL+"/api/token?grant_type=client_credentials", nil)
	if err != nil {
		return "", Error{"(*Client).getToken", "http.NewRequest", err}
	}

	req.Header.Set("Authorization", c.basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", Error{"(*Client).getToken", "(*http.Client).Do", err}
	}

	defer res.Body.Close()

	var data map[string]any
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", Error{"(*Client).getToken", "(*json.Decoder).Decode", err}
	}

	if _, ok := data["access_token"]; !ok {
		return "", Error{"(*Client).getToken", "(*json.Decoder).Decode", errors.New(res.Status + " \"access_token\" field does not exist in map of type map[string]any")}
	}

	return "Bearer " + data["access_token"].(string), nil
}

func (c *Client) getOAuth() (*OAuth, error) {
	values := url.Values{
		"client_id":     {c.id},
		"response_type": {"code"},
		"redirect_uri":  {"http://localhost:8888/callback"},
		"scope":         {"playlist-modify-public playlist-modify-private"},
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", ACC_URL+"/authorize?"+values.Encode())
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", ACC_URL+"/authorize?"+values.Encode())
	case "darwin":
		cmd = exec.Command("open", ACC_URL+"/authorize?"+values.Encode())
	default:
		return &OAuth{}, Error{"(*Client).getOAuth", "exec.Command", errors.New(runtime.GOOS + " is not a supported platform")}
	}

	if err := cmd.Start(); err != nil {
		return &OAuth{}, Error{"(*Client).getOAuth", "exec.Command", err}
	}

	var queries url.Values

	srv := &http.Server{Addr: ":8888"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		queries = r.URL.Query()
		http.ServeFile(w, r, "index.html")
		srv.Shutdown(context.Background())
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return &OAuth{}, Error{"(*Client).getOAuth", "(*http.Server).ListenAndServe", err}
	}

	if queries.Get("error") != "" {
		return &OAuth{}, Error{"(*Client).getOAuth", "(*http.Server).ListenAndServe", errors.New(queries.Get("error") + ": there was an issue with the authentication flow")}
	}

	values = url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {queries.Get("code")},
		"redirect_uri": {"http://localhost:8888/callback"},
	}

	req, err := http.NewRequest("POST", ACC_URL+"/api/token?"+values.Encode(), nil)
	if err != nil {
		return &OAuth{}, Error{"(*Client).getOAuth", "http.NewRequest", err}
	}

	req.Header.Set("Authorization", c.basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &OAuth{}, Error{"(*Client).getOAuth", "(*http.Client).Do", err}
	}

	defer res.Body.Close()

	var data map[string]any
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return &OAuth{}, Error{"(*Client).getOAuth", "(*json.Decoder).Decode", err}
	}

	if _, ok := data["access_token"]; !ok {
		return &OAuth{}, Error{"(*Client).getOAuth", "(*json.Decoder).Decode", errors.New(res.Status + " \"access_token\" field does not exist in map of type map[string]any")}
	}

	return &OAuth{token: "Bearer " + data["access_token"].(string), refresh: data["refresh_token"].(string)}, nil
}

// func (c *Client) refreshOAuth() (string, error) {
// 	values := url.Values{
// 		"grant_type":    {"refresh_token"},
// 		"refresh_token": {c.oauth.refresh},
// 	}

// 	req, err := http.NewRequest("POST", ACC_URL+"/api/token?"+values.Encode(), nil)
// 	if err != nil {
// 		return "", Error{"(*Client).refreshOAuth", "http.NewRequest", err}
// 	}

// 	req.Header.Set("Authorization", c.basic)
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return "", Error{"(*Client).refreshOAuth", "(*http.Client).Do", err}
// 	}

// 	defer res.Body.Close()

// 	var data map[string]any
// 	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
// 		return "", Error{"(*Client).refreshOAuth", "(*json.Decoder).Decode", err}
// 	}

// 	if _, ok := data["access_token"]; !ok {
// 		return "", Error{"(*Client).refreshOAuth", "(*json.Decoder).Decode", errors.New(res.Status + " \"access_token\" field does not exist in map of type map[string]any")}
// 	}

// 	return "Bearer " + data["access_token"].(string), nil
// }

func (c *Client) getUserID() (string, error) {
	req, err := http.NewRequest("GET", API_URL+"/me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.oauth.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var data map[string]any
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", Error{"(*Client).getUserID", "(*json.Decoder).Decode", err}
	}

	if _, ok := data["id"]; !ok {
		return "", Error{"(*Client).getUserID", "(*json.Decoder).Decode", errors.New(res.Status + " \"id\" field does not exist in map of type map[string]any")}
	}

	return data["id"].(string), nil
}
