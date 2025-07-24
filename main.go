package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"spotivert.com/spotivert/apple"
	"spotivert.com/spotivert/spotify"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

var (
	wg     sync.WaitGroup
	config Config
	logger zerolog.Logger
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

	log.Logger.Info().Msg("What is the URL for the Apple Music playlist you are trying to convert?\n")

	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	playlistURL := strings.TrimSpace(scanner.Text())

	wg.Add(1)
	var playlist apple.Playlist
	go func() {
		defer wg.Done()
		if playlist, err = apple.GetTracks(playlistURL); err != nil {
			logger.Fatal().Msg(err.Error())
		}
	}()

	log.Logger.Info().Msg("What would you like to name the Spotify playlist after converted?\n")

	scanner.Scan()
	playlistName := strings.TrimSpace(scanner.Text())

	wg.Wait()

	converted := make([]string, len(playlist.Songs))

	bar := progressbar.Default(int64(len(playlist.Songs)))

	for i, s := range playlist.Songs {
		wg.Add(1)
		go func(index int, song apple.Song) {
			defer wg.Done()
			defer bar.Add(1)

			res, err := spotifyClient.SearchTrack(song.Title, song.Artist, true)
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

	id, err := spotifyClient.CreatePlaylist(playlistName)
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
