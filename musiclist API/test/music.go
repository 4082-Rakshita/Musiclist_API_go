package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
)

// User represents a user in the MusicLister application.
type User struct {
	ID         string
	SecretCode string
	Name       string
	Email      string
	Playlists  []Playlist
}

// Playlist represents a playlist in the MusicLister applicataion.
type Playlist struct {
	ID     string
	Name   string
	Songs  []Song
	UserID string
}

// Song represents a song in the MusicLister application.
type Song struct {
	ID        string
	Name      string
	Composers string
	MusicURL  string
}

// MusicListerAPI is the main struct that holds the data and provides HTTP handlers.
type MusicListerAPI struct {
	Users     map[string]User
	Playlists map[string]Playlist
	Songs     map[string]Song
	Mutex     sync.RWMutex
}

// NewMusicListerAPI creates a new instance of the MusicListerAPI.
func NewMusicListerAPI() *MusicListerAPI {
	return &MusicListerAPI{
		Users:     make(map[string]User),
		Playlists: make(map[string]Playlist),
		Songs:     make(map[string]Song),
	}
}

// RegisterUser handles user registration.
func (api *MusicListerAPI) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.Name == "" || newUser.Email == "" {
		http.Error(w, "Name and Email are required", http.StatusBadRequest)
		return
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	// Check if user with the same email already exists
	for _, user := range api.Users {
		if user.Email == newUser.Email {
			http.Error(w, "User with this email already exists", http.StatusBadRequest)
			return
		}
	}

	newUser.ID = generateUniqueID()
	newUser.SecretCode = generateUniqueID()
	api.Users[newUser.SecretCode] = newUser

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// LoginUser handles user login.
func (api *MusicListerAPI) LoginUser(w http.ResponseWriter, r *http.Request) {
	secretCode := r.URL.Query().Get("secretCode")

	api.Mutex.RLock()
	defer api.Mutex.RUnlock()

	user, exists := api.Users[secretCode]
	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

// ViewProfile handles viewing user profiles.
func (api *MusicListerAPI) ViewProfile(w http.ResponseWriter, r *http.Request) {
	secretCode := r.URL.Query().Get("secretCode")

	api.Mutex.RLock()
	defer api.Mutex.RUnlock()

	user, exists := api.Users[secretCode]
	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

// GetAllSongsOfPlaylist handles getting all songs of a playlist.
func (api *MusicListerAPI) GetAllSongsOfPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID := r.URL.Query().Get("playlistId")

	api.Mutex.RLock()
	defer api.Mutex.RUnlock()

	playlist, exists := api.Playlists[playlistID]
	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(playlist.Songs)
	} else {
		http.Error(w, "Playlist not found", http.StatusNotFound)
	}
}

// CreatePlaylist handles creating a new playlist.
func (api *MusicListerAPI) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	secretCode := r.URL.Query().Get("secretCode")

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	user, exists := api.Users[secretCode]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var newPlaylist Playlist
	err := json.NewDecoder(r.Body).Decode(&newPlaylist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newPlaylist.ID = generateUniqueID()
	newPlaylist.UserID = user.ID
	api.Playlists[newPlaylist.ID] = newPlaylist

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPlaylist)
}

// DeletePlaylist handles deleting a playlist.
func (api *MusicListerAPI) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID := r.URL.Query().Get("playlistId")

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	_, exists := api.Playlists[playlistID]
	if exists {
		delete(api.Playlists, playlistID)
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Playlist not found", http.StatusNotFound)
	}
}

// GetSongDetail handles getting details of a song.
func (api *MusicListerAPI) GetSongDetail(w http.ResponseWriter, r *http.Request) {
	songID := r.URL.Query().Get("songId")

	api.Mutex.RLock()
	defer api.Mutex.RUnlock()

	song, exists := api.Songs[songID]
	if exists {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(song)
	} else {
		http.Error(w, "Song not found", http.StatusNotFound)
	}
}

// addSongToPlaylist handles adding a song to a playlist.
func (api *MusicListerAPI) addSongToPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID := r.URL.Query().Get("playlistId")

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	playlist, exists := api.Playlists[playlistID]
	if !exists {
		http.Error(w, "Playlist not found", http.StatusNotFound)
		return
	}

	var newSong Song
	err := json.NewDecoder(r.Body).Decode(&newSong)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newSong.ID = generateUniqueID()
	playlist.Songs = append(playlist.Songs, newSong)
	api.Playlists[playlistID] = playlist

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playlist)
}

// generateUniqueID generates a unique ID using the UUID library.
func generateUniqueID() string {
	id := uuid.New()
	return id.String()
}

func main() {
	api := NewMusicListerAPI()

	http.HandleFunc("/register", api.RegisterUser)
	http.HandleFunc("/login", api.LoginUser)
	http.HandleFunc("/ViewProfile", api.ViewProfile)
	http.HandleFunc("/getAllSongsOfPlaylist", api.GetAllSongsOfPlaylist)
	http.HandleFunc("/createPlaylist", api.CreatePlaylist)
	http.HandleFunc("/deletePlaylist", api.DeletePlaylist)
	http.HandleFunc("/getSongDetail", api.GetSongDetail)
	http.HandleFunc("/addSongToPlaylist", api.addSongToPlaylist)

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
