package helpers

import (
	"fmt"
	"html/template"
	"log/slog"
	"sort"
	"strings"

	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/index"
	"github.com/sokkalf/hubro/utils"
)

func TagCloudInit(idx *index.Index) {
	// This goroutine listens for signals generated by the index, and resets the cache when it receives one
	go func() {
		msgChan := idx.MsgBroker.Subscribe()
		for {
			switch <-msgChan {
			case index.Updated:
				slog.Debug("Resetting tag cloud cache")
				globalCache.reset()
			default: // Ignore other messages
			}
		}
	}()
}

func tagCloudMap(idx *index.Index) map[string]int {
	tagCloud := make(map[string]int)
	for _, entry := range idx.GetEntries() {
		for _, tag := range entry.Tags {
			tagCloud[tag]++
		}
	}
	return tagCloud
}

func GenerateTagCloud(idx *index.Index) template.HTML {
	if t, ok := globalCache.get(idx); ok {
		return *t
	}

	tagCloud := tagCloudMap(idx)
	var max int
	for _, count := range tagCloud {
		if count > max {
			max = count
		}
	}

	cssTextSizeClasses := []string{"text-xs", "text-sm", "text-base", "text-lg",
		"text-xl", "text-2xl", "text-3xl", "text-4xl"}
	cssTextSize := func(count int) string {
		count-- // 0-indexed, and count is guaranteed to be at least 1
		return cssTextSizeClasses[((count)*len(cssTextSizeClasses))/max]
	}

	tagHTML := func(tag string, count int) string {
		return fmt.Sprintf(
			`<span class="%s"><a data-hx-boost="true" href="%s?tag=%s">%s</a></span>%s`,
			cssTextSize(count),
			config.Config.RootPath,
			tag,
			tag,
			"\n",
		)
	}

	var sortedTags []string
	for tag := range tagCloud {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	tagCloudHTML := strings.Join(utils.Map(func(tag string) string {
		return tagHTML(tag, tagCloud[tag])
	}, sortedTags), "")

	tmpl := template.HTML(tagCloudHTML)
	globalCache.set(idx, &tmpl)
	return tmpl
}
