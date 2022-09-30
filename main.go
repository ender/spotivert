package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"spotivert/apple"
	"spotivert/spotify"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

var (
	wg     sync.WaitGroup
	config Config
	logger zerolog.Logger

	appleRegex = regexp.MustCompile(`https?://(?:itunes|music)\.apple\.com/.+?(album|playlist).*\/([\w\.\-]+)`)
)

type Config struct {
	Spotify struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	} `json:"spotify"`
}

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "03:04:05", NoColor: true}).With().Timestamp().Logger()

	if _, err := os.Stat("log"); os.IsNotExist(err) {
		err = os.Mkdir("log", 0777)
		if err != nil {
			log.Logger.Fatal().Msg(err.Error())
		}
	}

	logFile, err := os.OpenFile("log/spotivert-log-"+time.Now().Format("2006-01-02Z15-04-05"+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Logger.Fatal().Msg(err.Error())
	}

	logger = zerolog.New(zerolog.ConsoleWriter{Out: logFile, TimeFormat: "03:04:05", NoColor: true}).With().Timestamp().Logger()
	log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: io.MultiWriter(os.Stdout, logFile), TimeFormat: "03:04:05", NoColor: true}).With().Timestamp().Logger()

	configFile, err := os.Open("config.json")
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		logger.Fatal().Msg(err.Error())
	}
}

func main() {
	spotifyClient, err := spotify.New(config.Spotify.ID, config.Spotify.Secret)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := spotifyClient.Authenticate(); err != nil {
			logger.Fatal().Msg(err.Error())
		}
	}()

	wg.Add(1)
	var appleClient *apple.Client
	go func() {
		defer wg.Done()
		token, err := getAppleToken()
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
		appleClient = apple.New(token)
	}()

	log.Logger.Info().Msg("What is the URL for the Apple Music playlist you are trying to convert?\n")

	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	playlistURL := strings.TrimSpace(scanner.Text())

	log.Logger.Info().Msg("What would you like to name the Spotify playlist after converted?\n")

	scanner.Scan()
	playlistName := strings.TrimSpace(scanner.Text())

	matches := appleRegex.FindAllStringSubmatch(playlistURL, -1)
	typ, id := matches[0][1], matches[0][2]

	wg.Wait()

	appleSongs, err := appleClient.GetTracks(typ, id)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	converted := make([]string, len(appleSongs))

	wg.Wait()

	bar := progressbar.Default(int64(len(appleSongs)))
	for i, s := range appleSongs {
		wg.Add(1)
		go func(index int, song apple.Song) {
			defer wg.Done()
			defer bar.Add(1)

			res, err := spotifyClient.SearchTrack(song.Name, song.Artist, true)
			if err != nil {
				logger.Warn().Msg(err.Error())
				return
			}

			converted[index] = res.URI
		}(i, s)
	}

	wg.Wait()

	bar.Clear()

	var empty int
	for _, v := range converted {
		if v == "" {
			empty++
		}
	}

	if empty != 0 {
		log.Logger.Warn().Msg(strconv.Itoa(empty) + " songs were unable to be found. Check the log file for more information.")
	} else {
		log.Logger.Info().Msg("All songs have been successfully converted.")
	}

	log.Logger.Info().Msg("Now adding songs to user's Spotify playlist.")

	id, err = spotifyClient.CreatePlaylist(playlistName)
	if err != nil {
		log.Logger.Fatal().Msg(err.Error())
	}

	for _, batch := range splitArray(converted, 100) {
		err = spotifyClient.AddItems(id, batch)
		if err != nil {
			log.Logger.Warn().Msg(err.Error())
		}
	}

	log.Logger.Info().Msg("Songs added to playlist. Enjoy :)")
}

func getAppleToken() (string, error) {
	res, err := http.Get("https://music.apple.com")
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var query string
	if str, exists := doc.Find("meta[name='desktop-music-app/config/environment']").Attr("content"); exists {
		if query, err = url.QueryUnescape(str); err != nil {
			return "", err
		}
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(query), &data); err != nil {
		return "", err
	}

	return "Bearer " + data["MEDIA_API"].(map[string]any)["token"].(string), nil
}
