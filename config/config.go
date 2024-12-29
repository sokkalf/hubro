package config

import (
	"log/slog"
	"net/url"
	"os"
	"strconv"
)

type HubroConfig struct {
	BaseURL string
	Port    int
	AuthorName string
	AuthorEmail string
	FeedsEnabled bool
	DisplayAuthorInFeed bool
	Title string
	Description string
	RootPath string
	LegacyRoutesFile string
	BlogDir string
	PagesDir string
	Version string
}

var Config *HubroConfig

func Init() {
	config := HubroConfig{
		BaseURL: "http://localhost:8080/",
		Port: 8080,
		AuthorName: "Anonymous",
		AuthorEmail: "anonymous@example.org",
		FeedsEnabled: true,
		DisplayAuthorInFeed: false,
		Title: "Hubro",
		Description: "Hubro is a simple blog engine",
		LegacyRoutesFile: "./legacyRoutes.json",
		BlogDir: "./blog",
		PagesDir: "./pages",
		Version: "0.0.1-dev",
	}

	baseURL, ok := os.LookupEnv("HUBRO_BASE_URL")
	if ok {
		config.BaseURL = baseURL
	}
	path, err := url.Parse(config.BaseURL)
	if err != nil {
		slog.Error("Error parsing base URL", "error", err, "url", config.BaseURL)
	} else {
		var port int
		config.RootPath = path.Path
		if path.Scheme == "http" && path.Port() == "" {
			port = 80
		} else if path.Scheme == "https" && path.Port() == "" {
			port = 443
		} else {
			port, err = strconv.Atoi(path.Port())
			if err != nil {
				port = 8080
			}
		}
		config.Port = port
	}
	port, ok := os.LookupEnv("HUBRO_PORT")
	if ok {
		var err error
		config.Port, err = strconv.Atoi(port)
		if err != nil {
			config.Port = 8080
		}
	}
	authorName, ok := os.LookupEnv("HUBRO_AUTHOR_NAME")
	if ok {
		config.AuthorName = authorName
	}
	authorEmail, ok := os.LookupEnv("HUBRO_AUTHOR_EMAIL")
	if ok {
		config.AuthorEmail = authorEmail
	}
	title, ok := os.LookupEnv("HUBRO_TITLE")
	if ok {
		config.Title = title
	}
	description, ok := os.LookupEnv("HUBRO_DESCRIPTION")
	if ok {
		config.Description = description
	}
	displayAuthorInFeed, ok := os.LookupEnv("HUBRO_DISPLAY_AUTHOR_IN_FEED")
	if ok {
		config.DisplayAuthorInFeed, _ = strconv.ParseBool(displayAuthorInFeed)
	}
	feedsEnabled, ok := os.LookupEnv("HUBRO_FEEDS_ENABLED")
	if ok {
		config.FeedsEnabled, _ = strconv.ParseBool(feedsEnabled)
	}
	legacyRoutesFile, ok := os.LookupEnv("HUBRO_LEGACY_ROUTES_FILE")
	if ok {
		config.LegacyRoutesFile = legacyRoutesFile
	}
	blogDir, ok := os.LookupEnv("HUBRO_BLOG_DIR")
	if ok {
		config.BlogDir = blogDir
	}
	pagesDir, ok := os.LookupEnv("HUBRO_PAGES_DIR")
	if ok {
		config.PagesDir = pagesDir
	}
	Config = &config
}
