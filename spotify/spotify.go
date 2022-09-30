package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	API_URL = "https://api.spotify.com/v1"
	ACC_URL = "https://accounts.spotify.com"
)

type Client struct {
	id    string
	basic string
	token string
	oauth *OAuth
}

type Results struct {
	Data `json:"tracks"`
}

type Data struct {
	URL   string `json:"href"`
	Songs []Song `json:"items"`
}

type Song struct {
	URI     string `json:"uri"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Name       string `json:"name"`
	Popularity int    `json:"popularity"`
}

func New(id, secret string) (*Client, error) {
	c := &Client{
		id:    id,
		basic: "Basic " + base64.StdEncoding.EncodeToString([]byte(id+":"+secret)),
	}

	token, err := c.getToken()
	if err != nil {
		return nil, Error{"New", "(*Client).getToken", err}
	}

	c.token = token

	return c, nil
}

func (c *Client) SearchTrack(query, artist string, retry bool) (Song, error) {
	encoded := url.QueryEscape(sanitizeQuery(query) + ": " + strings.ReplaceAll(artist, "& ", ","))

	req, err := http.NewRequest("GET", API_URL+"/search?q="+encoded+"&type=track&limit=3", nil)
	if err != nil {
		return Song{}, Error{"(*Client).SearchTrack", "http.NewRequest", err}
	}

recurse:
	req.Header.Set("Authorization", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Song{}, Error{"(*Client).SearchTrack", "(*http.Client).Do", err}
	}

	if res.StatusCode == 401 || res.StatusCode == 429 {
		if token, _ := c.getToken(); token != "" {
			c.token = token
		}
		goto recurse
	}

	defer res.Body.Close()

	var results Results
	if err = json.NewDecoder(res.Body).Decode(&results); err != nil {
		return Song{}, Error{"(*Client).SearchTrack", "(*json.Decoder).Decode", err}
	}

	if len(results.Songs) == 0 {
		if !retry {
			return Song{}, Error{"(*Client).SearchTrack", "(*http.Client).Do", errors.New(res.Status + " " + "could not find track \"" + sanitizeQuery(query) + ": " + artist + "\"")}
		}
		return c.SearchTrack(query, artist, false)
	}

	last := Song{Popularity: 0}
	for _, song := range results.Songs {
		if strings.EqualFold(sanitizeString(song.Name), sanitizeString(query)) {
			if song.Popularity > last.Popularity {
				last = song
			}
		}
	}

	if last.Name != "" {
		return last, nil
	}

	return results.Songs[0], nil
}

func (c *Client) CreatePlaylist(name string) (string, error) {
	req, err := http.NewRequest("POST", API_URL+"/users/"+c.oauth.me+"/playlists", bytes.NewBuffer([]byte("{\"name\": \""+name+"\"}")))
	if err != nil {
		return "", Error{"(*Client).CreatePlaylist", "http.NewRequest", err}
	}

	req.Header.Set("Authorization", c.oauth.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", Error{"(*Client).CreatePlaylist", "(*http.Client).Do", err}
	}

	if res.StatusCode == 429 {
		return c.CreatePlaylist(name)
	}

	defer res.Body.Close()

	// var data map[string]any
	// if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
	// 	return "", Error{"(*Client).CreatePlaylist", "(*json.Decoder).Decode", err}
	// }

	body, _ := io.ReadAll(res.Body)

	var data map[string]any
	if err = json.Unmarshal(body, &data); err != nil {
		fmt.Println(string(body))
		return "", Error{"(*Client).CreatePlaylist", "(*json.Decoder).Decode", err}
	}

	if _, ok := data["id"]; !ok {
		return "", Error{"(*Client).CreatePlaylist", "(*json.Decoder).Decode", errors.New(res.Status + " \"id\" field does not exist in map of type map[string]any")}
	}

	return data["id"].(string), nil
}

func (c *Client) AddItems(playlist string, items []string) error {
	data, err := json.Marshal(map[string][]string{"uris": items})
	if err != nil {
		return Error{"(*Client).AddItems", "json.Marshal", err}
	}

	req, err := http.NewRequest("POST", API_URL+"/playlists/"+playlist+"/tracks", bytes.NewBuffer(data))
	if err != nil {
		return Error{"(*Client).AddItems", "http.NewRequest", err}
	}

	req.Header.Set("Authorization", c.oauth.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Error{"(*Client).AddItems", "(*http.Client).Do", err}
	}

	if res.StatusCode != 201 {
		time.Sleep(2 * time.Second)
		return c.AddItems(playlist, items)
	}

	return nil
}
