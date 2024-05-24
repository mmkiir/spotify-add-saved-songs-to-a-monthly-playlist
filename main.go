package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// ReadTokenFromPath reads a token from a file at the given path.
func ReadTokenFromPath(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// GetPlaylist retrieves a playlist by its ID.
func GetPlaylist(
	client *http.Client,
	playlistId string,
) (map[string]interface{}, error) {
	resp, err := client.Get(
		fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", playlistId),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var playlist map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	return playlist, nil
}

// GetCurrentUsersProfile retrieves the profile of the current user.
func GetCurrentUsersProfile(
	client *http.Client,
) (map[string]interface{}, error) {
	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		return nil, err
	}

	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// EnumerateCurrentUsersPlaylists retrieves all playlists of the current user.
func EnumerateCurrentUsersPlaylists(
	client *http.Client,
) ([]interface{}, error) {
	resp, err := client.Get("https://api.spotify.com/v1/me/playlists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var playlists map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return nil, err
	}

	items, ok := playlists["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"unexpected type for items: %T",
			playlists["items"],
		)
	}

	next, ok := playlists["next"]
	if !ok {
		return nil, fmt.Errorf(
			"unexpected type for next: %T",
			playlists["next"],
		)
	}

	for next != nil {
		resp, err := client.Get(next.(string))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
			return nil, err
		}

		nextItems, ok := playlists["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"unexpected type for items: %T",
				playlists["items"],
			)
		}

		items = append(items, nextItems...)

		next, ok = playlists["next"]
		if !ok {
			return nil, fmt.Errorf(
				"unexpected type for next: %T",
				playlists["next"],
			)
		}
	}

	return items, nil
}

// EnumerateUsersSavedTracks retrieves all saved tracks of the current user.
func EnumerateUsersSavedTracks(client *http.Client) ([]interface{}, error) {
	resp, err := client.Get("https://api.spotify.com/v1/me/tracks")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tracks map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, err
	}

	items, ok := tracks["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for items: %T", tracks["items"])
	}

	next, ok := tracks["next"]
	if !ok {
		return nil, fmt.Errorf("unexpected type for next: %T", tracks["next"])
	}

	for next != nil {
		resp, err := client.Get(next.(string))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
			return nil, err
		}

		nextItems, ok := tracks["items"].([]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"unexpected type for items: %T",
				tracks["items"],
			)
		}

		items = append(items, nextItems...)

		next, ok = tracks["next"]
		if !ok {
			return nil, fmt.Errorf(
				"unexpected type for next: %T",
				tracks["next"],
			)
		}
	}

	return items, nil
}

// AddItemsToPlaylist adds items to a playlist.
func AddItemsToPlaylist(
	client *http.Client,
	playlistId string,
	uris []string,
) error {
	body := map[string]interface{}{
		"uris": uris,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"https://api.spotify.com/v1/playlists/%s/tracks",
			playlistId,
		),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CreatePlaylist creates a playlist.
func CreatePlaylist(
	client *http.Client,
	userId string,
	name string,
	public bool,
	collaborative bool,
	description string,
) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"name":          name,
		"public":        public,
		"collaborative": collaborative,
		"description":   description,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userId),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var playlist map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	return playlist, nil
}

// WriteTokenToPath writes a token to a file at the given path.
func WriteTokenToPath(path string, token *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return err
	}

	return nil
}

// DeletePlaylist deletes a playlist by its ID.
func DeletePlaylist(client *http.Client, playlistId string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(
			"https://api.spotify.com/v1/playlists/%s/followers",
			playlistId,
		),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"failed to delete playlist, status code: %d",
			resp.StatusCode,
		)
	}

	return nil
}

// DeletePlaylistsByNameFormat deletes playlists by a regex pattern.
func DeletePlaylistsByNameFormat(
	client *http.Client,
	regexPattern string,
) error {
	playlists, err := EnumerateCurrentUsersPlaylists(client)
	if err != nil {
		return err
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return err
	}

	for _, playlist := range playlists {
		plMap, ok := playlist.(map[string]interface{})
		if !ok {
			log.Fatalf("unexpected type for playlist: %T", playlist)
		}

		playlistName, ok := plMap["name"].(string)
		if !ok {
			log.Printf("unexpected type for name: %T", plMap["name"])
		}

		if regex.MatchString(playlistName) {
			log.Printf("deleting playlist: %s", playlistName)
			playlistId, ok := plMap["id"].(string)
			if !ok {
				log.Printf("unexpected type for id: %T", plMap["id"])
			}

			if err := DeletePlaylist(client, playlistId); err != nil {
				return err
			}
		}
	}

	return nil
}

// Path: main.go
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err)
	}

	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectUri := os.Getenv("SPOTIFY_REDIRECT_URI")

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: redirectUri,
		Scopes: []string{
			"playlist-modify-private",
			"playlist-modify-public",
			"playlist-read-private",
			"user-library-read",
			"user-read-private",
			"user-read-email",
		},
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}

	path := filepath.Join(
		dir,
		"spotify-add-saved-songs-to-a-monthly-playlist",
		"token.json",
	)

	token, err := ReadTokenFromPath(path)
	if err != nil {
		url := config.AuthCodeURL("state")
		log.Println("Visit the URL for the auth dialog:", url)
		log.Println("Enter the code:")

		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatalln(err)
		}

		token, err = config.Exchange(ctx, code)
		if err != nil {
			log.Fatalln(err)
		}

		if err := WriteTokenToPath(path, token); err != nil {
			log.Fatalln(err)
		}
	}

	tokenSource := config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln(err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := WriteTokenToPath(path, newToken); err != nil {
			log.Fatalln(err)
		}
	}

	client := config.Client(ctx, newToken)

	// if err := DeletePlaylistsByNameFormat(client, `^[A-Za-z]+\s'\d{2}$`); err != nil {
	// 	log.Fatalln(err)
	// }

	profile, err := GetCurrentUsersProfile(client)
	if err != nil {
		log.Fatalln(err)
	}

	id, ok := profile["id"].(string)
	if !ok {
		log.Fatalf("unexpected type for id: %T", profile["id"])
	}

	playlists, err := EnumerateCurrentUsersPlaylists(client)
	if err != nil {
		log.Fatalln(err)
	}

	tracks, err := EnumerateUsersSavedTracks(client)
	if err != nil {
		log.Fatalln(err)
	}

	reverseTracks := make([]interface{}, len(tracks))
	for i, j := 0, len(tracks)-1; i < j; i, j = i+1, j-1 {
		reverseTracks[i], reverseTracks[j] = tracks[j], tracks[i]
	}

	// Map to keep track of created/target playlists for each month
	playlistMap := make(map[string]string)
	tracksByMonth := make(map[string][]string)

	for _, track := range reverseTracks {
		trackMap, ok := track.(map[string]interface{})
		if !ok {
			log.Fatalf("unexpected type for track: %T", track)
		}

		addedAt, ok := trackMap["added_at"].(string)
		if !ok {
			log.Fatalf("unexpected type for added_at: %T", trackMap["added_at"])
		}

		trackMap = trackMap["track"].(map[string]interface{})
		if !ok {
			log.Fatalf("unexpected type for track: %T", trackMap["track"])
		}

		t, err := time.Parse(time.RFC3339, addedAt)
		if err != nil {
			log.Fatalln(err)
		}

		targetPlaylistName := t.Format("January '06")
		uri := trackMap["uri"].(string)
		tracksByMonth[targetPlaylistName] = append(
			tracksByMonth[targetPlaylistName],
			uri,
		)

		if _, exists := playlistMap[targetPlaylistName]; !exists {
			for _, playlist := range playlists {
				plMap, ok := playlist.(map[string]interface{})
				if !ok {
					log.Fatalf("unexpected type for playlist: %T", playlist)
				}

				playlistName, ok := plMap["name"].(string)
				if !ok {
					log.Printf(
						"unexpected type for playlistName: %T",
						plMap["name"],
					)
				}

				if playlistName == targetPlaylistName {
					log.Printf("found %s", targetPlaylistName)
					playlistMap[targetPlaylistName] = plMap["id"].(string)
					break
				}
			}

			if _, exists := playlistMap[targetPlaylistName]; !exists {
				log.Printf("creating %s", targetPlaylistName)
				playlist, err := CreatePlaylist(
					client,
					id,
					targetPlaylistName,
					true,
					false,
					"",
				)
				if err != nil {
					log.Fatalln(err)
				}

				playlistMap[targetPlaylistName] = playlist["id"].(string)
			}
		}
	}

	for month, uris := range tracksByMonth {
		if len(uris) > 0 {
			targetPlaylistId := playlistMap[month]

			playlist, err := GetPlaylist(client, targetPlaylistId)
			if err != nil {
				log.Fatalln(err)
			}

			existingTracks := make(map[string]bool)
			tracks, ok := playlist["tracks"].(map[string]interface{})
			if !ok {
				log.Fatalf("unexpected type for tracks: %T", playlist["tracks"])
			}

			items, ok := tracks["items"].([]interface{})
			if !ok {
				log.Fatalf("unexpected type for items: %T", tracks["items"])
			}

			for _, item := range items {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					log.Fatalf("unexpected type for item: %T", item)
				}

				trackMap, ok := itemMap["track"].(map[string]interface{})
				if !ok {
					log.Fatalf(
						"unexpected type for track: %T",
						itemMap["track"],
					)
				}

				uri := trackMap["uri"].(string)
				existingTracks[uri] = true
			}

			newUris := []string{}
			for _, uri := range uris {
				if !existingTracks[uri] {
					newUris = append(newUris, uri)
				}
			}

			if len(newUris) > 0 {
				if err := AddItemsToPlaylist(client, targetPlaylistId, newUris); err != nil {
					log.Fatalln(err)
				}
				log.Printf("added %d tracks to %s", len(newUris), month)
			}
		}
	}
}
