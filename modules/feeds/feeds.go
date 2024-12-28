package feeds

import (
	"log/slog"
	"net/http"
	"time"

	gorillafeeds "github.com/gorilla/feeds"
	"github.com/sokkalf/hubro/server"
)

func getFeedFromIndex(index *server.Index) *gorillafeeds.Feed {
	feed := &gorillafeeds.Feed{
		Title:       "Hubro",
		Link:        &gorillafeeds.Link{Href: "http://localhost:8080"},
		Description: "Hubro is a simple blog engine",
		Author:      &gorillafeeds.Author{Name: "Christian Lønaas", Email: "email@example.org"},
		Created:	 time.Now(),
	}

	feedItems := []*gorillafeeds.Item{}
	for _, entry := range index.Entries {
		var summary string
		if entry.Summary != nil {
			summary = string(*entry.Summary)
		} else {
			summary = "Description not available"
		}

		feedItems = append(feedItems, &gorillafeeds.Item{
			Title:       entry.Title,
			Link:        &gorillafeeds.Link{Href: "http://localhost:8080" + entry.Path},
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
