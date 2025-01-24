package config

import (
	"log/slog"
	"net/url"
	"os"
	"strconv"
)

type HubroConfig struct {
	BaseURL             string
	Port                int
	AuthorName          string
	AuthorEmail         string
	FeedsEnabled        bool
	DisplayAuthorInFeed bool
	Title               string
	Description         string
	RootPath            string
	LegacyRoutesFile    string
	BlogDir             string
	PagesDir            string
	UserStaticDir       string
	LogoImage           string
	UserCSS             bool
	PostsPerPage        int
	Version             string
	Environment         string
	GelfEndpoint        *string
	AdminEnabled        bool
	AdminPassword       string
}

var Config *HubroConfig

func Init() {
	config := HubroConfig{
		BaseURL:             "http://localhost:8080/",
		Port:                8080,
		AuthorName:          "Anonymous",
		AuthorEmail:         "anonymous@example.org",
		FeedsEnabled:        true,
		DisplayAuthorInFeed: false,
		Title:               "Hubro",
		Description:         "Hubro is a simple blog engine",
		LegacyRoutesFile:    "./legacyRoutes.json",
		BlogDir:             "./blog",
		PagesDir:            "./pages",
		UserStaticDir:       "./userfiles",
		LogoImage:           "logo.svg",
		PostsPerPage:        10,
		Version:             "0.0.1-dev",
		Environment:         "development",
		GelfEndpoint:        nil,
	}

	if baseURL, ok := os.LookupEnv("HUBRO_BASE_URL"); ok {
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
	if port, ok := os.LookupEnv("HUBRO_PORT"); ok {
		var err error
		config.Port, err = strconv.Atoi(port)
		if err != nil {
			config.Port = 8080
		}
	}
	if authorName, ok := os.LookupEnv("HUBRO_AUTHOR_NAME"); ok {
		config.AuthorName = authorName
	}
	if authorEmail, ok := os.LookupEnv("HUBRO_AUTHOR_EMAIL"); ok {
		config.AuthorEmail = authorEmail
	}
	if title, ok := os.LookupEnv("HUBRO_TITLE"); ok {
		config.Title = title
	}
	if description, ok := os.LookupEnv("HUBRO_DESCRIPTION"); ok {
		config.Description = description
	}
	if displayAuthorInFeed, ok := os.LookupEnv("HUBRO_DISPLAY_AUTHOR_IN_FEED"); ok {
		config.DisplayAuthorInFeed, _ = strconv.ParseBool(displayAuthorInFeed)
	}
	if feedsEnabled, ok := os.LookupEnv("HUBRO_FEEDS_ENABLED"); ok {
		config.FeedsEnabled, _ = strconv.ParseBool(feedsEnabled)
	}
	if legacyRoutesFile, ok := os.LookupEnv("HUBRO_LEGACY_ROUTES_FILE"); ok {
		config.LegacyRoutesFile = legacyRoutesFile
	}
	if blogDir, ok := os.LookupEnv("HUBRO_BLOG_DIR"); ok {
		config.BlogDir = blogDir
	}
	if pagesDir, ok := os.LookupEnv("HUBRO_PAGES_DIR"); ok {
		config.PagesDir = pagesDir
	}
	if environment, ok := os.LookupEnv("HUBRO_ENVIRONMENT"); ok {
		config.Environment = environment
	}
	if adminEnabled, ok := os.LookupEnv("HUBRO_ADMIN_ENABLED"); ok {
		config.AdminEnabled, _ = strconv.ParseBool(adminEnabled)
	}
	if adminPassword, ok := os.LookupEnv("HUBRO_ADMIN_PASSWORD"); ok {
		config.AdminPassword = adminPassword
	} else {
		if config.AdminEnabled {
			config.AdminEnabled = false
			slog.Warn("Admin interface disabled, no password set")
		}
	}
	if gelfEndpoint, ok := os.LookupEnv("HUBRO_GELF_ENDPOINT"); ok {
		config.GelfEndpoint = &gelfEndpoint
	}
	if userStaticDir, ok := os.LookupEnv("HUBRO_USERFILES_DIR"); ok {
		config.UserStaticDir = userStaticDir
	}
	if logoImage, ok := os.LookupEnv("HUBRO_LOGO_IMAGE"); ok {
		config.LogoImage = logoImage
	}
	fi, err := os.Stat(config.UserStaticDir + "/" + config.LogoImage)
	if err != nil && fi == nil {
		config.LogoImage = "" // no logo image found
	}
	fi, err = os.Stat(config.UserStaticDir + "/user.css")
	if err != nil && fi == nil {
		config.UserCSS = false
	} else {
		config.UserCSS = true
	}
	Config = &config
}
