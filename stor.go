package main

import (
	"encoding/json"
	log "github.com/pymhd/go-logging"
	"os"
	"time"
	"sync"
)

const (
	DefStorFile = "/tmp/mshows.cache"
)

var (
	sm ShowsManager
)

func init() {
	sm = ShowsManager{}
	sm.LoadItself()
}

type ShowsManager struct {
	mu        sync.Mutex
	Token     string       `json:"token"`
	Episodes  map[int]bool `json:"episodes"`
	SentItems map[int]bool `json:"sent"`
}

func (s *ShowsManager) AddToWatchlist(ids ...int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.Episodes[id] = true
	}
}

func (s *ShowsManager) DelFromWatchlist(ids ...int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		s.Episodes[id] = false
	}
}

func (s *ShowsManager) IsMonitored(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Episodes[id]
}

func (s *ShowsManager) MarkAsSent(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SentItems[id] = true
}

func (s *ShowsManager) IsSent(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.SentItems[id]
}

func (s *ShowsManager) Save() {
	now := time.Now()
	defer func() {
        	log.Debugf("Storing data to disk took: %v\n", time.Since(now))
	}()
	
	f, err := os.Create(DefStorFile)
	if err != nil {
		log.Errorf("Save file failed: %s\n", err)
		return
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(s); err != nil {
		log.Errorf("Encode data to file failed: %s\n", err)
	}
}

func (s *ShowsManager) LoadItself() {
	f, _ := os.Open(DefStorFile)
	defer f.Close()

	err := json.NewDecoder(f).Decode(s)
	if err != nil {
		log.Warningf("Load file failed: %s\n", err)
		sm.SentItems = make(map[int]bool, 0)
		sm.Episodes = make(map[int]bool, 0)
		log.Debugln("Empty cached initialized")
	}
}
