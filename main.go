package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andybrewer/mack"
)

// GitLabIssue represents the structure of a GitLab issue
type GitLabIssue struct {
	ID          int       `json:"id"`
	IID         int       `json:"iid"`
	ProjectID   int       `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DueDate     string    `json:"due_date"`
	WebURL      string    `json:"web_url"`
	Labels      []string  `json:"labels"`
}

// Config represents the application configuration
type Config struct {
	GitLabToken    string `json:"gitlab_token"`
	GitLabURL      string `json:"gitlab_url"`
	GitLabUsername string `json:"gitlab_username"`
	ReminderList   string `json:"reminder_list"`
	PollInterval   int    `json:"poll_interval_minutes"`
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func getAssignedIssues(config Config) ([]GitLabIssue, error) {
	url := fmt.Sprintf("%s/api/v4/issues?assignee_username=%s&state=opened",
		config.GitLabURL, config.GitLabUsername)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", config.GitLabToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var issues []GitLabIssue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func reminderExists(title string, list string) (bool, error) {
	// Check if reminder with the same title already exists in the specified list
	script := fmt.Sprintf(`
		tell application "Reminders"
			set myList to list "%s"
			set matchingReminders to (reminders of myList whose name is "%s")
			if (count of matchingReminders) > 0 then
				return "true"
			else
				return "false"
			end if
		end tell
	`, list, escapeAppleScriptString(title))

	result, err := mack.Tell("Reminders", script)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(result) == "true", nil
}

func createReminder(issue GitLabIssue, list string) error {
	title := fmt.Sprintf("#%d: %s", issue.IID, issue.Title)
	notes := fmt.Sprintf("%s\n\nURL: %s", issue.Description, issue.WebURL)

	// Check if reminder already exists
	exists, err := reminderExists(title, list)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("Reminder already exists for issue #%d", issue.IID)
		return nil
	}

	script := fmt.Sprintf(`
		tell application "Reminders"
			tell list "%s"
				make new reminder with properties {name:"%s", body:"%s"}
			end tell
		end tell
	`, list, escapeAppleScriptString(title), escapeAppleScriptString(notes))

	_, err = mack.Tell("Reminders", script)
	if err != nil {
		return err
	}

	log.Printf("Created reminder for issue #%d: %s", issue.IID, issue.Title)
	return nil
}

func escapeAppleScriptString(s string) string {
	// Replace double quotes with escaped double quotes for AppleScript
	return strings.ReplaceAll(s, `"`, `\"`)
}

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Printf("Starting GitLab to Apple Reminders integration")
	log.Printf("Monitoring issues assigned to: %s", config.GitLabUsername)
	log.Printf("Creating reminders in list: %s", config.ReminderList)

	// Initial sync
	syncIssues(config)

	// Set up periodic syncing
	ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		syncIssues(config)
	}
}

func syncIssues(config Config) {
	issues, err := getAssignedIssues(config)
	if err != nil {
		log.Printf("Error fetching GitLab issues: %v", err)
		return
	}

	log.Printf("Found %d assigned issues", len(issues))

	for _, issue := range issues {
		err := createReminder(issue, config.ReminderList)
		if err != nil {
			log.Printf("Error creating reminder for issue #%d: %v", issue.IID, err)
		}
	}
}
