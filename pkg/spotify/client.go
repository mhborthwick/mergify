package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

type Track struct {
	URI string `json:"uri"`
}

type PlaylistTrack struct {
	Track Track `json:"track"`
}

type PlaylistItemsResponse struct {
	Items []PlaylistTrack `json:"items"`
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
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
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

/*
TODO: Spotify API constrains you to 20 playlists
per request. Implement some kind of paginated playlists
retrieval logic in case you have more than 20 playlists.
*/
func (s *Spotify) getPlaylists(userID string) ([]Playlist, error) {
	endpoint := fmt.Sprintf("/users/%s/playlists", userID)
	body, err := s.handleRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var response PlaylistsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return response.Items, nil
}

/*
GetPlaylistIDsByName retrieves the IDs corresponding
to the playlists provided in the user's mergify config.
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

/*
TODO: Handle duplicate IDs.
*/
func (s *Spotify) GetPlaylistTrackIDs(playlistIDs []string) ([]string, error) {
	var result []string
	for _, id := range playlistIDs {
		playlistTracks, err := s.GetTracksFromPlaylist(id)
		if err != nil {
			return nil, err
		}
		for _, p := range playlistTracks {
			result = append(result, p.Track.URI)
		}
	}
	return result, nil
}

/*
TODO: Spotify API constrains you to 20 tracks
per request. Implement some kind of paginated track
retrieval logic in case you have more than 20 tracks.
*/
func (s *Spotify) GetTracksFromPlaylist(playlistID string) ([]PlaylistTrack, error) {
	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	body, err := s.handleRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var response PlaylistItemsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return response.Items, nil
}

/*
TODO: Create only if trackURIs length > 0
TODO: Change 'name' in payload
*/
func (s *Spotify) CreatePlaylist(userID string) (string, error) {
	requestBody := map[string]string{"name": "Mergify Playlist"}
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
		return "", fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return response.ID, nil
}

/*
TODO: Spotify API constrains you to 100 URIs
per request. Send tracks in batches of 100.
*/
func (s *Spotify) AddTracksToPlaylist(playlistID string, trackIDs []string) (string, error) {
	requestBody := map[string][]string{"uris": trackIDs}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	body, err := s.handleRequest("POST", endpoint, bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return "", err
	}
	var response AddTracksToPlaylistResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return response.SnapshotID, nil
}
