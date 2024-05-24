# Spotify Add Saved Songs to a Monthly Playlist

This script will add all of your saved songs to a playlist for the current month. It will also remove any duplicates that are already in the playlist.

## Prerequisites

- Go 1.16 or later
- A Spotify developer account
- Spotify API credentials (client ID, client secret and redirect URI)

## Installation

1. Clone the repository

```bash
git clone https://github.com/mmkiir/spotify-add-saved-songs-to-monthly-playlist.git
```

2. Change directory

```bash
cd spotify-add-saved-songs-to-monthly-playlist
```

3. Install dependencies

```bash
go mod tidy
```

4. Create a `.env` file in the root of the project and add the following environment variables:

```bash
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
SPOTIFY_REDIRECT_URI=your_spotify_redirect_uri
```

## Usage

1. Run the script

```bash
go run main.go
```

2. Open the URL in your browser and authorize the application

3. Copy the `code` parameter from the URL and paste it in the terminal

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
