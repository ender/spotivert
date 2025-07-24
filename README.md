# 🎵 Spotivert

**Spotivert** is a lightweight Go tool that converts **Apple Music playlists** to **Spotify playlists** by guiding you through a simple interactive prompt. It uses web scraping along with the Spotify API to find matching songs and automatically recreate your playlist on Spotify.

> ✅ Apple Music → Spotify conversion supported  
> 🔜 Spotify → Apple Music and other improvements planned  
> 💡 Interactive, user-friendly prompt-based interface

---

## 🚀 Features

- 🧭 Interactive prompts guide the entire process
- ✅ Transfers full playlists from Apple Music to Spotify
- 🔗 Matches songs using title, artist, and duration
- 🧱 Built in Go — portable, fast, and dependency-light

---

## 📦 Installation

Make sure you have **Go 1.20+** installed.

```bash
git clone https://github.com/your-username/spotivert.git
cd spotivert
go build -o spotivert
```

---

## 🧰 Requirements

- A **Spotify Developer account**: [Spotify Developer Portal](https://developer.spotify.com/)
- An Apple Music playlist URL

---

## ⚙️ Usage

Run the program and follow the prompt instructions:

```bash
./spotivert
```

You’ll be asked to:

1. Paste your Apple Music playlist URL
2. Authenticate with your Spotify account
3. Confirm playlist details and track matches

---

## 📋 TODO

- [ ] Add **Spotify → Apple Music** conversion
- [ ] Implement **local caching** to avoid re-fetching and re-matching the same songs
- [ ] Support **single-command mode** with flags (e.g. `spotivert --from apple --to spotify`)
- [ ] Add support for **ISRC-based** matching for improved accuracy
- [ ] Optional **web-based UI** for broader accessibility

---

## 🛠 Tech Stack

- Language: **Go**
- APIs: **Spotify Web API**
- Interface: **Prompt-based CLI**
- Planned: JSON file or embedded DB for caching

---

## 🔗 Resources

- [Spotify Developer Portal](https://developer.spotify.com/)
