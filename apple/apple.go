package apple

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

type results []struct {
	Song   string `json:"song"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

type Playlist struct {
	Songs []Song
}

type Song struct {
	Title  string
	Artist string
	Album  string
}

func GetTracks(url string) (Playlist, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var results results

	// JavaScript that scrapes the song details
	script := `
		[...document.querySelectorAll('.songs-list-row__song-name-wrapper')].map(wrapper => {
			const songEl = wrapper.querySelector('.songs-list-row__song-name');
			const artistEl = wrapper.querySelector('.songs-list-row__by-line a');
			const row = wrapper.closest('.songs-list-row'); 
			let albumEl = null;
			if(row) {
				albumEl = row.querySelector('.songs-list__col--tertiary a');
			}

			return {
				song: songEl ? songEl.textContent.trim() : "",
				artist: artistEl ? artistEl.textContent.trim() : "",
				album: albumEl ? albumEl.textContent.trim() : ""
			};
		})
	`

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.songs-list-row__song-name-wrapper`, chromedp.ByQuery),
	}

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		return Playlist{}, Error{"GetTracks", "chromedp.Run", err}
	}

	// Loop to scroll and wait for the page to fully load
	for {
		var isFooterVisible bool
		if err = chromedp.Run(ctx, chromedp.EvaluateAsDevTools(`document.querySelector('[data-testid="tracklist-footer-description"]') !== null;`, &isFooterVisible)); err != nil {
			return Playlist{}, Error{"GetTracks", "chromedp.Run", err}
		}

		if isFooterVisible {
			break
		}

		if err = chromedp.Run(ctx, chromedp.KeyEvent(kb.End)); err != nil {
			return Playlist{}, Error{"GetTracks", "chromedp.KeyEvent", err}
		}

		time.Sleep(200 * time.Millisecond)
	}

	err = chromedp.Run(ctx, chromedp.Evaluate(script, &results))
	if err != nil {
		return Playlist{}, Error{"GetTracks", "chromedp.Evaluate", err}
	}

	playlist := Playlist{Songs: []Song{}}
	for _, item := range results {
		song := Song{
			item.Song,
			item.Artist,
			item.Album,
		}
		playlist.Songs = append(playlist.Songs, song)
	}

	return playlist, nil
}
