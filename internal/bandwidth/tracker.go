package bandwidth

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"multi-protocol-proxy/internal/ui"
)

type UserUsage struct {
	BytesUp      int64  `json:"bytes_up"`
	BytesDown    int64  `json:"bytes_down"`
	TotalBytes   int64  `json:"total_bytes"`
	LastResetAt  string `json:"last_reset_at"`
	ActiveConns  int    `json:"active_conns"`
}

type UsageFile struct {
	Month string                `json:"month"` 
	Users map[string]*UserUsage `json:"users"`
}

type Tracker struct {
	mu       sync.Mutex
	users    map[string]*UserUsage
	month    string 
	filePath string
	stopCh   chan struct{}
}

func NewTracker(filePath string) *Tracker {
	t := &Tracker{
		users:    make(map[string]*UserUsage),
		month:    time.Now().Format("2006-01"),
		filePath: filePath,
		stopCh:   make(chan struct{}),
	}

	t.loadFromDisk()

	go t.backgroundLoop()

	return t
}

func (t *Tracker) RecordBytes(username string, up, down int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.checkMonthlyReset()

	u := t.getOrCreate(username)
	u.BytesUp += up
	u.BytesDown += down
	u.TotalBytes += up + down
}


func (t *Tracker) CheckAllowance(username string, limitGB int) bool {
	if limitGB <= 0 {
		return true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.checkMonthlyReset()

	u := t.getOrCreate(username)
	limitBytes := int64(limitGB) * 1024 * 1024 * 1024
	return u.TotalBytes < limitBytes
}

func (t *Tracker) GetUsage(username string) UserUsage {
	t.mu.Lock()
	defer t.mu.Unlock()

	u := t.getOrCreate(username)
	return *u
}

func (t *Tracker) GetAllUsage() map[string]UserUsage {
	t.mu.Lock()
	defer t.mu.Unlock()

	result := make(map[string]UserUsage, len(t.users))
	for k, v := range t.users {
		result[k] = *v
	}
	return result
}

func (t *Tracker) IncrementConns(username string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	u := t.getOrCreate(username)
	u.ActiveConns++
}

func (t *Tracker) DecrementConns(username string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	u := t.getOrCreate(username)
	if u.ActiveConns > 0 {
		u.ActiveConns--
	}
}

func (t *Tracker) GetActiveConns(username string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	u := t.getOrCreate(username)
	return u.ActiveConns
}


func (t *Tracker) CheckConnLimit(username string, maxConns int) bool {
	if maxConns <= 0 {
		return true
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	u := t.getOrCreate(username)
	return u.ActiveConns < maxConns
}

func (t *Tracker) Stop() {
	close(t.stopCh)
	t.saveToDisk() 
}

func (t *Tracker) GetMonth() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.month
}


func (t *Tracker) getOrCreate(username string) *UserUsage {
	u, ok := t.users[username]
	if !ok {
		u = &UserUsage{
			LastResetAt: time.Now().Format(time.RFC3339),
		}
		t.users[username] = u
	}
	return u
}

func (t *Tracker) checkMonthlyReset() {
	currentMonth := time.Now().Format("2006-01")
	if currentMonth != t.month {
		ui.LogStatus("info", fmt.Sprintf("Monthly bandwidth reset: %s → %s", t.month, currentMonth))
		for _, u := range t.users {
			u.BytesUp = 0
			u.BytesDown = 0
			u.TotalBytes = 0
			u.LastResetAt = time.Now().Format(time.RFC3339)
		}
		t.month = currentMonth
		t.saveToDiskLocked()
	}
}

func (t *Tracker) backgroundLoop() {
	saveTicker := time.NewTicker(5 * time.Minute)
	defer saveTicker.Stop()

	for {
		select {
		case <-saveTicker.C:
			t.saveToDisk()
		case <-t.stopCh:
			return
		}
	}
}

func (t *Tracker) saveToDisk() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.saveToDiskLocked()
}

func (t *Tracker) saveToDiskLocked() {
	file := UsageFile{
		Month: t.month,
		Users: t.users,
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		ui.LogStatus("error", "Failed to marshal bandwidth usage: "+err.Error())
		return
	}
	if err := os.WriteFile(t.filePath, data, 0644); err != nil {
		ui.LogStatus("error", "Failed to save bandwidth usage: "+err.Error())
	}
}

func (t *Tracker) loadFromDisk() {
	data, err := os.ReadFile(t.filePath)
	if err != nil {
		return
	}

	var file UsageFile
	if err := json.Unmarshal(data, &file); err != nil {
		ui.LogStatus("warn", "Failed to parse bandwidth usage file, starting fresh: "+err.Error())
		return
	}

	currentMonth := time.Now().Format("2006-01")
	if file.Month == currentMonth && file.Users != nil {
		t.users = file.Users
		t.month = file.Month
		for _, u := range t.users {
			u.ActiveConns = 0
		}
		ui.LogStatus("info", fmt.Sprintf("Restored bandwidth usage for %d users (month: %s)", len(t.users), t.month))
	} else {
		ui.LogStatus("info", fmt.Sprintf("Bandwidth data from %s discarded (current month: %s)", file.Month, currentMonth))
	}
}
