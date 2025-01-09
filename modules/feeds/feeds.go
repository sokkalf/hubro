package feeds

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	gorillafeeds "github.com/gorilla/feeds"
	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/server"
)

type Feeds struct {
	feedCache      map[*index.Index]*gorillafeeds.Feed
	feedCacheMutex sync.RWMutex
}

func InitFeeds(i *index.Index) *Feeds {
	f := &Feeds{
		feedCache: make(map[*index.Index]*gorillafeeds.Feed),
	}
	f.feedCache[i] = getFeedFromIndex(i)

	go func() {
		msgChan := i.MsgBroker.Subscribe()
		for {
			switch <-msgChan {
			case index.Updated:
				slog.Debug("Resetting feed cache")
				f.feedCacheMutex.Lock()
				f.feedCache[i] = getFeedFromIndex(i)
				f.feedCacheMutex.Unlock()
			default: // Ignore other messages
			}
		}
	}()
	return f
}

func getFeedFromIndex(index *index.Index) *gorillafeeds.Feed {
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
			Description: entry.Description,
			Created:     entry.Date,
			Content:     summary,
		})
	}
	feed.Items = feedItems
	return feed
}

func Register(prefix string, h *server.Hubro, mux *http.ServeMux, options interface{}) {
	start := time.Now()
	index := options.(*index.Index)
	feeds := InitFeeds(index)
	if feeds == nil {
		slog.Error("Failed to initialize feeds")
		return
	}

	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		feeds.feedCacheMutex.RLock()
		feeds.feedCache[index].WriteRss(w)
		feeds.feedCacheMutex.RUnlock()
	})
	mux.HandleFunc("/atom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		feeds.feedCacheMutex.RLock()
		feeds.feedCache[index].WriteAtom(w)
		feeds.feedCacheMutex.RUnlock()
	})
	slog.Info("Registered feeds", "atomUrl", prefix+"/atom", "rssUrl", prefix+"/rss", "duration", time.Since(start))
}
