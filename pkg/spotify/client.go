package spotify

import (
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
	ID string `json:"id"`
}

type PlaylistTrack struct {
	Track Track `json:"track"`
}

type PlaylistItemsResponse struct {
	Items []PlaylistTrack `json:"items"`
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
	if resp.StatusCode != http.StatusOK {
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

func (s *Spotify) GetPlaylistIDsByName(userID string, cfgPlaylists []string) ([]string, error) {
	playlists, err := s.GetPlaylists(userID)
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
TODO: Spotify API constrains you to 100 playlists
per request. Implement some kind of paginated playlists
retrieval logic in case you have more than 100 playlists.
*/
func (s *Spotify) GetPlaylists(userID string) ([]Playlist, error) {
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
			result = append(result, p.Track.ID)
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
