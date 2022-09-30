package apple

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

const API_URL = "https://api.music.apple.com"

type Client struct {
	token string
}

type Playlist struct {
	Next  string `json:"next"`
	Songs []Song `json:"data"`

	Errors []struct {
		Message string `json:"detail"`
	} `json:"errors"`
}

type Song struct {
	Attributes `json:"attributes"`
}

type Attributes struct {
	Name   string `json:"name"`
	Artist string `json:"artistName"`
}

func New(token string) *Client {
	return &Client{
		token: token,
	}
}

func (c *Client) GetTracks(typ, id string) (songs []Song, err error) {
	req, err := http.NewRequest("GET", API_URL+"/v1/catalog/us/"+typ+"s/"+id+"/tracks?limit=300", nil)
	if err != nil {
		return []Song{}, Error{"(*Client).GetTracks", "http.NewRequest", err}
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("referer", "https://music.apple.com")
	req.Header.Set("origin", "https://music.apple.com")

recurse:
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Song{}, Error{"(*Client).GetTracks", "(*http.Client).Do", err}
	}

	defer res.Body.Close()

	var playlist Playlist
	err = json.NewDecoder(res.Body).Decode(&playlist)
	if err != nil {
		return []Song{}, Error{"(*Client).GetTracks", "(*json.Decoder).Decode", err}
	}

	if res.StatusCode != 200 {
		return []Song{}, Error{"(*Client).GetTracks", "(http.Response).StatusCode", errors.New(res.Status + " " + playlist.Errors[0].Message)}
	}

	songs = append(songs, playlist.Songs...)

	if playlist.Next != "" {
		req.URL, err = url.Parse(API_URL + playlist.Next + "&limit=300")
		if err != nil {
			return []Song{}, Error{"(*Client).GetTracks", "url.Parse", err}
		}
		goto recurse
	}

	return songs, nil
}
