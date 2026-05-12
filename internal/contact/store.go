package contact

import (
    "bufio"
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "sort"
    "sync"
)

type Store interface { Save(Message) error; List() ([]Message, error); UpdateStatus(string, string) (Message, bool, error); Delete(string) (bool, error) }
type FileStore struct { path string; mu sync.Mutex }
func NewFileStore(path string) *FileStore { return &FileStore{path: path} }

func (s *FileStore) Save(message Message) error {
    s.mu.Lock(); defer s.mu.Unlock()
    if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil { return err }
    file, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); if err != nil { return err }
    defer file.Close()
    encoded, err := json.Marshal(message); if err != nil { return err }
    _, err = file.Write(append(encoded, '\n')); return err
}

func (s *FileStore) List() ([]Message, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.readAllLocked() }

func (s *FileStore) UpdateStatus(id string, status string) (Message, bool, error) {
    s.mu.Lock(); defer s.mu.Unlock()
    messages, err := s.readAllLocked(); if err != nil { return Message{}, false, err }
    for i := range messages { if messages[i].ID == id { messages[i].Status = status; if err := s.writeAllLocked(messages); err != nil { return Message{}, false, err }; return messages[i], true, nil } }
    return Message{}, false, nil
}

func (s *FileStore) Delete(id string) (bool, error) {
    s.mu.Lock(); defer s.mu.Unlock()
    messages, err := s.readAllLocked(); if err != nil { return false, err }
    next := messages[:0]; found := false
    for _, message := range messages { if message.ID == id { found = true; continue }; next = append(next, message) }
    if !found { return false, nil }
    return true, s.writeAllLocked(next)
}

func (s *FileStore) readAllLocked() ([]Message, error) {
    file, err := os.Open(s.path); if errors.Is(err, os.ErrNotExist) { return []Message{}, nil }; if err != nil { return nil, err }
    defer file.Close()
    messages := make([]Message, 0); scanner := bufio.NewScanner(file)
    for scanner.Scan() { var message Message; if err := json.Unmarshal(scanner.Bytes(), &message); err != nil { return nil, err }; messages = append(messages, message) }
    if err := scanner.Err(); err != nil { return nil, err }
    sort.Slice(messages, func(i, j int) bool { return messages[i].CreatedAt.After(messages[j].CreatedAt) })
    return messages, nil
}

func (s *FileStore) writeAllLocked(messages []Message) error {
    if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil { return err }
    tmp := s.path + ".tmp"; file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); if err != nil { return err }
    for _, message := range messages { encoded, err := json.Marshal(message); if err != nil { file.Close(); return err }; if _, err := file.Write(append(encoded, '\n')); err != nil { file.Close(); return err } }
    if err := file.Close(); err != nil { return err }
    return os.Rename(tmp, s.path)
}
