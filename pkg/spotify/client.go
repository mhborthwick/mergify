package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Spotify struct {
	Token  string
	Client *http.Client
	UserID string
}

type Profile struct {
	ID string `json:"id"`
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PlaylistsResponse struct {
	Items []Playlist `json:"items"`
	Next  *string    `json:"next"`
}

type Track struct {
	URI string `json:"uri"`
}

type PlaylistTrack struct {
	Track Track `json:"track"`
}

type PlaylistItemsResponse struct {
	Items []PlaylistTrack `json:"items"`
	Next  *string         `json:"next"`
}

type AddTracksToPlaylistResponse struct {
	SnapshotID string `json:"snapshot_id"`
}

type CreatePlaylistResponse struct {
	ID string `json:"id"`
}

const API = "https://api.spotify.com/v1"

func (s *Spotify) handleRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	if s.Client == nil {
		s.Client = &http.Client{}
	}
	url := API + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if method == "GET" && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	if method == "POST" && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return respBody, nil
}

func (s *Spotify) getProfile() (*Profile, error) {
	body, err := s.handleRequest("GET", "/me", nil)
	if err != nil {
		return nil, err
	}
	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &profile, nil
}

func (s *Spotify) GetUserID() (string, error) {
	profile, err := s.getProfile()
	if err != nil {
		return "", err
	}
	return profile.ID, nil
}

func (s *Spotify) getPlaylists(userID string) ([]Playlist, error) {
	var allPlaylists []Playlist
	endpoint := fmt.Sprintf("/users/%s/playlists", userID)
	/*
		Spotify imposes a 20 playlists per request constraint, so we need
		to add logic to be able to retrieve playlists in multiple cycles.
	*/
	for {
		body, err := s.handleRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		var response PlaylistsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal playlists: %w", err)
		}
		allPlaylists = append(allPlaylists, response.Items...)
		if response.Next == nil {
			break
		}
		_, after, _ := strings.Cut(*response.Next, API)
		endpoint = after
	}
	return allPlaylists, nil
}

/*
GetPlaylistIDsByName retrieves the IDs corresponding
to the playlists provided in the user's ~/.mergify/config.json file.
*/
func (s *Spotify) GetPlaylistIDsByName(userID string, cfgPlaylists []string) ([]string, error) {
	playlists, err := s.getPlaylists(userID)
	if err != nil {
		return nil, err
	}
	hashMap := make(map[string]string)
	for _, playlist := range playlists {
		hashMap[playlist.Name] = playlist.ID
	}
	var result []string
	for _, name := range cfgPlaylists {
		if id, exists := hashMap[name]; exists {
			result = append(result, id)
		}
	}
	return result, nil
}

func (s *Spotify) GetPlaylistTrackIDs(playlistIDs []string) ([]string, error) {
	var trackURIs []string
	hashMap := make(map[string]bool)
	for _, id := range playlistIDs {
		playlistTracks, err := s.getTracksFromPlaylist(id)
		if err != nil {
			return nil, err
		}
		/*
			Omits duplicate Track IDs
			to prevent the created playlist
			from having duplicate tracks.
		*/
		for _, p := range playlistTracks {
			if !hashMap[p.Track.URI] {
				trackURIs = append(trackURIs, p.Track.URI)
				hashMap[p.Track.URI] = true
			}
		}
	}
	return trackURIs, nil
}

func (s *Spotify) getTracksFromPlaylist(playlistID string) ([]PlaylistTrack, error) {
	var allPlaylistTracks []PlaylistTrack
	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	/*
		Spotify defaults to returning 20 tracks
		per request, so we need to implement track retrieval mechanism
		that can handle playlists with more than 20 tracks.
	*/
	for {
		body, err := s.handleRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		var response PlaylistItemsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		allPlaylistTracks = append(allPlaylistTracks, response.Items...)
		if response.Next == nil {
			break
		}
		_, after, _ := strings.Cut(*response.Next, API)
		endpoint = after
	}
	return allPlaylistTracks, nil
}

func (s *Spotify) CreatePlaylist(userID string, trackIDs []string) (string, error) {
	if len(trackIDs) == 0 {
		/*
			Exit if no tracks found in playlists
			provided in user's ~/.mergify/config.json file.
		*/
		return "", fmt.Errorf("no tracks found")
	}
	now := time.Now()
	millis := now.UnixNano() / 1e6
	name := fmt.Sprintf("Mergify Playlist %d", millis)
	requestBody := map[string]string{
		"name":        name,
		"description": "Created with https://github.com/mhborthwick/mergify",
	}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("/users/%s/playlists", userID)
	body, err := s.handleRequest("POST", endpoint, bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return "", err
	}
	var response CreatePlaylistResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return response.ID, nil
}

func (s *Spotify) AddTracksToPlaylist(
	playlistID string,
	trackIDs []string,
	batchSize int,
) (string, error) {
	// Spotify limits you to max 100 URIs per request
	// so we need to be able to send tracks in batches.
	chunks := func(trackIDs []string, size int) [][]string {
		var result [][]string
		for i := 0; i < len(trackIDs); i += size {
			end := i + size
			if end > len(trackIDs) {
				end = len(trackIDs)
			}
			result = append(result, trackIDs[i:end])
		}
		return result
	}
	batches := chunks(trackIDs, batchSize)
	var lastSnapshotID string
	for _, batch := range batches {
		requestBody := map[string][]string{"uris": batch}
		jsonRequestBody, err := json.Marshal(requestBody)
		if err != nil {
			return "", err
		}
		endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)
		body, err := s.handleRequest("POST", endpoint, bytes.NewBuffer(jsonRequestBody))
		if err != nil {
			return "", fmt.Errorf("failed to add tracks to playlist: %w", err)
		}
		var response AddTracksToPlaylistResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}
		lastSnapshotID = response.SnapshotID
	}
	return lastSnapshotID, nil
}
