// +build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username           string `json:"username"`
	Role               string `json:"role"`
	PasswordHash       string `json:"password_hash"`
	RateLimitRPM       int    `json:"rate_limit_rpm"`
	Enabled            bool   `json:"enabled"`
	Plan               string `json:"plan,omitempty"`
	BandwidthLimitGB   int    `json:"bandwidth_limit_gb,omitempty"`
	BandwidthSpeedMbps int    `json:"bandwidth_speed_mbps,omitempty"`
	MaxConnections     int    `json:"max_connections,omitempty"`
	ExpiresAt          string `json:"expires_at,omitempty"`
}

type UsersConfig struct {
	Users         []User   `json:"users"`
	IPWhitelist   []string `json:"ip_whitelist"`
	SuperAdminIPs []string `json:"super_admin_ips,omitempty"`
}

var reader = bufio.NewReader(os.Stdin)

func main() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║   Multi-Protocol Proxy - User Management Tool║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	usersFile := "users.json"
	if len(os.Args) > 1 {
		usersFile = os.Args[1]
	}

	for {
		fmt.Println("What would you like to do?")
		fmt.Println("  1) Add a new user")
		fmt.Println("  2) List all users")
		fmt.Println("  3) Remove a user")
		fmt.Println("  4) Toggle user (enable/disable)")
		fmt.Println("  5) Update user plan/limits")
		fmt.Println("  6) Generate password hash only")
		fmt.Println("  7) Save and exit")
		fmt.Println("  8) Exit without saving")
		fmt.Println()

		choice := prompt("Choose an option (1-8)")
		cfg := loadConfig(usersFile)

		switch choice {
		case "1":
			addUser(cfg)
			saveConfig(usersFile, cfg)
		case "2":
			listUsers(cfg)
		case "3":
			removeUser(cfg)
			saveConfig(usersFile, cfg)
		case "4":
			toggleUser(cfg)
			saveConfig(usersFile, cfg)
		case "5":
			updateUser(cfg)
			saveConfig(usersFile, cfg)
		case "6":
			generateHash()
		case "7":
			saveConfig(usersFile, cfg)
			fmt.Println("\n✅ Saved to " + usersFile)
			fmt.Println("📋 Remember to copy this file to your server:")
			fmt.Println("   scp -i your-key.pem " + usersFile + " ubuntu@YOUR_EC2_IP:~")
			fmt.Println("   Then SSH in and run:")
			fmt.Println("   sudo mv ~/" + usersFile + " /opt/proxy/users.json")
			fmt.Println("   sudo chown proxy:proxy /opt/proxy/users.json")
			fmt.Println("   sudo systemctl restart proxy")
			return
		case "8":
			fmt.Println("Exiting without saving.")
			return
		default:
			fmt.Println("Invalid option. Try again.")
		}
		fmt.Println()
	}
}

func addUser(cfg *UsersConfig) {
	fmt.Println("\n── Add New User ──────────────────────────")

	username := prompt("Username")
	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}

	for _, u := range cfg.Users {
		if strings.EqualFold(u.Username, username) {
			fmt.Println("⚠️  User '" + username + "' already exists.")
			return
		}
	}

	password := prompt("Password")
	if password == "" {
		fmt.Println("Password cannot be empty.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}

	role := promptDefault("Role (user/super_admin)", "user")
	plan := promptDefault("Plan name (starter/pro/enterprise/admin)", "starter")

	bwLimitStr := promptDefault("Monthly bandwidth limit in GB (0 = unlimited)", "0")
	bwLimit, _ := strconv.Atoi(bwLimitStr)

	speedStr := promptDefault("Speed limit in Mbps (0 = unlimited, no throttle)", "0")
	speed, _ := strconv.Atoi(speedStr)

	maxConnsStr := promptDefault("Max concurrent connections (0 = unlimited)", "0")
	maxConns, _ := strconv.Atoi(maxConnsStr)

	rpmStr := promptDefault("Rate limit RPM - requests per minute (0 = unlimited)", "0")
	rpm, _ := strconv.Atoi(rpmStr)

	expiresAt := ""
	expiresInput := promptDefault("Expiry date - YYYY-MM-DD (empty = no expiry)", "")
	if expiresInput != "" {
		t, err := time.Parse("2006-01-02", expiresInput)
		if err != nil {
			fmt.Println("⚠️  Invalid date format, setting no expiry.")
		} else {
			expiresAt = t.Format(time.RFC3339)
		}
	}

	user := User{
		Username:           username,
		Role:               role,
		PasswordHash:       string(hash),
		RateLimitRPM:       rpm,
		Enabled:            true,
		Plan:               plan,
		BandwidthLimitGB:   bwLimit,
		BandwidthSpeedMbps: speed,
		MaxConnections:     maxConns,
		ExpiresAt:          expiresAt,
	}

	cfg.Users = append(cfg.Users, user)
	fmt.Println("\n✅ User '" + username + "' added successfully!")
	fmt.Println("   Plan: " + plan)
	if bwLimit > 0 {
		fmt.Println("   Bandwidth: " + bwLimitStr + " GB/month")
	} else {
		fmt.Println("   Bandwidth: Unlimited")
	}
	if speed > 0 {
		fmt.Println("   Speed: " + speedStr + " Mbps")
	} else {
		fmt.Println("   Speed: Unlimited")
	}
}

func listUsers(cfg *UsersConfig) {
	fmt.Println("\n── Current Users ─────────────────────────")
	if len(cfg.Users) == 0 {
		fmt.Println("  No users configured.")
		return
	}

	for i, u := range cfg.Users {
		status := "✅"
		if !u.Enabled {
			status = "❌"
		}
		fmt.Printf("  %d) %s %s [%s] plan=%s", i+1, status, u.Username, u.Role, u.Plan)

		if u.BandwidthLimitGB > 0 {
			fmt.Printf(" bw=%dGB", u.BandwidthLimitGB)
		} else {
			fmt.Print(" bw=∞")
		}

		if u.BandwidthSpeedMbps > 0 {
			fmt.Printf(" speed=%dMbps", u.BandwidthSpeedMbps)
		}

		if u.MaxConnections > 0 {
			fmt.Printf(" conns=%d", u.MaxConnections)
		}

		if u.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, u.ExpiresAt)
			if err == nil {
				if time.Now().After(t) {
					fmt.Print(" EXPIRED")
				} else {
					fmt.Printf(" expires=%s", t.Format("2006-01-02"))
				}
			}
		}

		fmt.Println()
	}
}

func removeUser(cfg *UsersConfig) {
	listUsers(cfg)
	if len(cfg.Users) == 0 {
		return
	}

	idxStr := prompt("Enter user number to remove (or 'cancel')")
	if idxStr == "cancel" {
		return
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(cfg.Users) {
		fmt.Println("Invalid selection.")
		return
	}

	removed := cfg.Users[idx-1]
	cfg.Users = append(cfg.Users[:idx-1], cfg.Users[idx:]...)
	fmt.Println("✅ Removed user: " + removed.Username)
}

func toggleUser(cfg *UsersConfig) {
	listUsers(cfg)
	if len(cfg.Users) == 0 {
		return
	}

	idxStr := prompt("Enter user number to toggle (or 'cancel')")
	if idxStr == "cancel" {
		return
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(cfg.Users) {
		fmt.Println("Invalid selection.")
		return
	}

	cfg.Users[idx-1].Enabled = !cfg.Users[idx-1].Enabled
	status := "enabled"
	if !cfg.Users[idx-1].Enabled {
		status = "disabled"
	}
	fmt.Println("✅ User '" + cfg.Users[idx-1].Username + "' is now " + status)
}

func updateUser(cfg *UsersConfig) {
	listUsers(cfg)
	if len(cfg.Users) == 0 {
		return
	}

	idxStr := prompt("Enter user number to update (or 'cancel')")
	if idxStr == "cancel" {
		return
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(cfg.Users) {
		fmt.Println("Invalid selection.")
		return
	}

	u := &cfg.Users[idx-1]
	fmt.Println("\nUpdating user: " + u.Username)
	fmt.Println("Press Enter to keep current value.\n")

	if v := promptDefault("Plan (current: "+u.Plan+")", ""); v != "" {
		u.Plan = v
	}

	if v := promptDefault("Bandwidth limit GB (current: "+strconv.Itoa(u.BandwidthLimitGB)+")", ""); v != "" {
		u.BandwidthLimitGB, _ = strconv.Atoi(v)
	}

	if v := promptDefault("Speed limit Mbps (current: "+strconv.Itoa(u.BandwidthSpeedMbps)+")", ""); v != "" {
		u.BandwidthSpeedMbps, _ = strconv.Atoi(v)
	}

	if v := promptDefault("Max connections (current: "+strconv.Itoa(u.MaxConnections)+")", ""); v != "" {
		u.MaxConnections, _ = strconv.Atoi(v)
	}

	if v := promptDefault("Rate limit RPM (current: "+strconv.Itoa(u.RateLimitRPM)+")", ""); v != "" {
		u.RateLimitRPM, _ = strconv.Atoi(v)
	}

	if v := promptDefault("New password (empty = keep current)", ""); v != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("Error hashing password:", err)
		} else {
			u.PasswordHash = string(hash)
			fmt.Println("Password updated!")
		}
	}

	if v := promptDefault("Expiry date YYYY-MM-DD (current: "+u.ExpiresAt+", 'clear' to remove)", ""); v != "" {
		if v == "clear" {
			u.ExpiresAt = ""
		} else {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				fmt.Println("⚠️  Invalid date format.")
			} else {
				u.ExpiresAt = t.Format(time.RFC3339)
			}
		}
	}

	fmt.Println("✅ User '" + u.Username + "' updated!")
}

func generateHash() {
	password := prompt("Enter password to hash")
	if password == "" {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nPassword hash (copy to users.json):")
	fmt.Println(string(hash))
}


func loadConfig(path string) *UsersConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return &UsersConfig{
			Users:       []User{},
			IPWhitelist: []string{},
		}
	}

	var cfg UsersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("⚠️  Warning: Could not parse " + path + ", starting fresh.")
		return &UsersConfig{
			Users:       []User{},
			IPWhitelist: []string{},
		}
	}

	return &cfg
}

func saveConfig(path string, cfg *UsersConfig) {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		fmt.Println("Error saving:", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func prompt(label string) string {
	fmt.Print(label + ": ")
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptDefault(label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Print(label + " [" + defaultValue + "]: ")
	} else {
		fmt.Print(label + ": ")
	}
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultValue
	}
	return text
}
