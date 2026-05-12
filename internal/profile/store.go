package profile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store interface {
	Get() (Profile, error)
	Save(Profile) (Profile, error)
}
type FileStore struct {
	path string
	mu   sync.Mutex
}

func NewFileStore(path string) *FileStore { return &FileStore{path: path} }

func (s *FileStore) Get() (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		profile := defaultProfile()
		return profile, s.writeLocked(profile)
	}
	if err != nil {
		return Profile{}, err
	}
	defer file.Close()
	var profile Profile
	if err := json.NewDecoder(file).Decode(&profile); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func (s *FileStore) Save(profile Profile) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return profile, s.writeLocked(profile)
}

func (s *FileStore) writeLocked(profile Profile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(profile)
}

func defaultProfile() Profile {
	return Profile{
		Name: "Babu Angi", Title: "SaaS Engineer and Full Stack Developer", Location: "Bangalore, India", Email: "contact@babuangi.com", Website: "https://babuangi.com",
		Summary:    "Full Stack Engineer building scalable SaaS platforms, AI-powered systems, admin dashboards, Flutter apps, and production-grade backend architecture.",
		Highlights: []Metric{{Label: "Experience", Value: "4+ years"}, {Label: "Projects", Value: "15+ shipped"}, {Label: "Focus", Value: "SaaS + AI"}},
		Links:      []ProfileLink{{Label: "Website", URL: "https://babuangi.com"}, {Label: "GitHub", URL: "https://github.com/babuangi"}, {Label: "LinkedIn", URL: "https://www.linkedin.com/in/babuangi/"}},
		FocusAreas: []string{"SaaS platforms", "Go backends", "Flutter apps", "AI integrations", "Admin dashboards"},
	}
}
