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
	Title string
	Description string
	RootPath string
}

var Config *HubroConfig

func Init() {
	config := HubroConfig{
		BaseURL: "http://localhost:8080/",
		Port: 8080,
		AuthorName: "Anonymous",
		AuthorEmail: "anonymous@example.org",
		Title: "Hubro",
		Description: "Hubro is a simple blog engine",
	}

	baseURL, ok := os.LookupEnv("HUBRO_BASE_URL")
	if ok {
		config.BaseURL = baseURL
	}
	path, err := url.Parse(config.BaseURL)
	if err != nil {
		slog.Error("Error parsing base URL", "error", err, "url", config.BaseURL)
	} else {
		config.RootPath = path.Path
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
	Config = &config
}
