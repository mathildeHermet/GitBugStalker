package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	repoURL              string
	discordWebhookURL    string
	googleChatWebhookURL string
	checkInterval        time.Duration
	cacheFile            string
	rateLimitDelay       = 500 * time.Millisecond // Delay between successive notifications
	lastChecked          time.Time
	logger               *log.Logger
)

func main() {
	// Define command-line flags
	flag.StringVar(&repoURL, "github-repo", "", "GitHub repository URL")
	flag.StringVar(&discordWebhookURL, "discord-hook-url", "", "Discord webhook URL")
	flag.StringVar(&googleChatWebhookURL, "google-chat-hook-url", "", "Google Chat webhook URL")
	flag.DurationVar(&checkInterval, "refresh-interval", 10*time.Minute, "Refresh interval")
	flag.StringVar(&cacheFile, "cache-file", ".local/cache/issue_cache.txt", "Cache file")
	flag.Parse()

	// Check that at least one webhook URL is provided
	if repoURL == "" || (discordWebhookURL == "" && googleChatWebhookURL == "") {
		log.Fatalf("GitHub repo and at least one webhook URL (--discord-hook-url or --google-chat-hook-url) are required")
	}

	// Initialize the logger to write to stdout
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	lastChecked = time.Now()

	for {
		checkForNewIssues()
		time.Sleep(checkInterval)
	}
}

func checkForNewIssues() {
	logger.Println("Checking for new issues...")
	resp, err := http.Get(repoURL)
	if err != nil {
		logger.Printf("Error fetching issues: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("Error fetching issues: received status code %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Error reading response body: %v\n", err)
		return
	}

	logger.Println("Parsing issues...")
	parseIssues(string(body))
}

func parseIssues(html string) {
	issueRegexp := regexp.MustCompile(`<a id="issue_(\d+)_link" class="Link--primary v-align-middle no-underline h4 js-navigation-open markdown-title" [^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := issueRegexp.FindAllStringSubmatch(html, -1)

	logger.Printf("Found %d matches\n", len(matches))

	// Load the cache
	cache, err := loadCache(cacheFile)
	if err != nil {
		logger.Printf("Error loading cache: %v\n", err)
		return
	}

	for _, match := range matches {
		issueID := match[1]
		issueURL := fmt.Sprintf("https://github.com%s", match[2])
		issueTitle := match[3]

		// Check if the issue ID is already in the cache
		if cache[issueID] {
			logger.Printf("Issue already notified: ID=%s, Title=%s, URL=%s\n", issueID, issueTitle, issueURL)
			continue
		}

		logger.Printf("Found new issue: ID=%s, Title=%s, URL=%s\n", issueID, issueTitle, issueURL)
		notify(issueTitle, issueURL)

		// Add the issue ID to the cache
		cache[issueID] = true

		// Delay to respect rate limits
		time.Sleep(rateLimitDelay)
	}

	// Save the updated cache
	if err := saveCache(cacheFile, cache); err != nil {
		logger.Printf("Error saving cache: %v\n", err)
	}
}

func notify(issueTitle, issueURL string) {
	if discordWebhookURL != "" {
		notifyDiscord(issueTitle, issueURL)
	}
	if googleChatWebhookURL != "" {
		notifyGoogleChat(issueTitle, issueURL)
	}
}

func notifyDiscord(issueTitle, issueURL string) {
	message := fmt.Sprintf("New Issue Created:\n\nTitle: %s\nURL: %s", issueTitle, issueURL)
	payload := map[string]string{"content": message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Printf("Error marshaling payload: %v\n", err)
		return
	}

	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Printf("Error sending to Discord: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.Printf("Non-204 response from Discord: %s\n", resp.Status)
	} else {
		logger.Println("Message successfully sent to Discord")
	}
}

func notifyGoogleChat(issueTitle, issueURL string) {
	message := fmt.Sprintf("New Issue Created:\n\nTitle: %s\nURL: %s", issueTitle, issueURL)
	payload := map[string]string{"text": message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Printf("Error marshaling payload: %v\n", err)
		return
	}

	resp, err := http.Post(googleChatWebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Printf("Error sending to Google Chat: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("Non-200 response from Google Chat: %s\n", resp.Status)
	} else {
		logger.Println("Message successfully sent to Google Chat")
	}
}

func loadCache(filename string) (map[string]bool, error) {
	cache := make(map[string]bool)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil // No cache file exists yet
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line != "" {
			cache[line] = true
		}
	}

	return cache, nil
}

func saveCache(filename string, cache map[string]bool) error {
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var lines []string
	for id := range cache {
		lines = append(lines, id)
	}

	data := strings.Join(lines, "\n")
	return os.WriteFile(filename, []byte(data), 0644)
}
