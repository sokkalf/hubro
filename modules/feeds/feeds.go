package feeds

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	gorillafeeds "github.com/gorilla/feeds"
	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/server"
)

func getFeedFromIndex(index *server.Index) *gorillafeeds.Feed {
	config := config.Config
	var author *gorillafeeds.Author
	if config.DisplayAuthorInFeed {
		author = &gorillafeeds.Author{Name: config.AuthorName, Email: config.AuthorEmail}
	} else {
		author = nil
	}
	feed := &gorillafeeds.Feed{
		Title:       "Hubro",
		Link:        &gorillafeeds.Link{Href: config.BaseURL},
		Description: config.Description,
		Author:      author,
		Created:     index.Entries[0].Date,
	}

	feedItems := []*gorillafeeds.Item{}
	for _, entry := range index.Entries {
		var summary string
		if entry.Summary != nil {
			summary = string(*entry.Summary)
		} else {
			summary = "Description not available"
		}
		baseURL := strings.TrimSuffix(config.BaseURL, "/")

		feedItems = append(feedItems, &gorillafeeds.Item{
			Title:       entry.Title,
			Link:        &gorillafeeds.Link{Href: baseURL + entry.Path},
			Description: summary,
			Created:     entry.Date,
			Content:     summary,
		})
	}
	feed.Items = feedItems
	return feed
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	index := options.(*server.Index)
	feed := getFeedFromIndex(index)

	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		feed.WriteRss(w)
	})
	mux.HandleFunc("/atom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		feed.WriteAtom(w)
	})
	slog.Info("Registered feeds", "atomUrl", prefix+"/atom", "rssUrl", prefix+"/rss", "duration", time.Since(start))
}
