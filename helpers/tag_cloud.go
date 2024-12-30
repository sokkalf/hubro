package helpers

import (
	"fmt"
	"html/template"
	"sort"

	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/index"
)

var generatedTagCloudMap map[*index.Index]*template.HTML

func tagCloudMap(idx *index.Index) map[string]int {
	tagCloud := make(map[string]int)
	for _, entry := range idx.Entries {
		for _, tag := range entry.Tags {
			tagCloud[tag]++
		}
	}
	return tagCloud
}

func GenerateTagCloud(idx *index.Index) template.HTML {
	if generatedTagCloudMap == nil {
		generatedTagCloudMap = make(map[*index.Index]*template.HTML)
	}
	if generatedTagCloudMap[idx] != nil {
		return *generatedTagCloudMap[idx]
	}
	tagCloud := tagCloudMap(idx)
	max := 0
	for _, count := range tagCloud {
		if count > max {
			max = count
		}
	}
	cssTextSizeClasses := []string{"text-xs", "text-sm", "text-base", "text-lg", "text-xl"}
	cssTextSize := func(count int) string {
		if count == 0 {
			return cssTextSizeClasses[0]
		}
		return cssTextSizeClasses[((count-1)*len(cssTextSizeClasses))/max]
	}

	tagHTML := func(tag string, count int) string {
		return fmt.Sprintf(`<span class="%s"><a hx-boost="true" href="%s?tag=%s">%s</a></span>%s`,
			cssTextSize(count), config.Config.RootPath, tag, tag, "\n")
	}

    var sortedTags []string
    for tag := range tagCloud {
        sortedTags = append(sortedTags, tag)
    }
    sort.Strings(sortedTags)

	var tagCloudHTML string
    for _, tag := range sortedTags {
        tagCloudHTML += tagHTML(tag, tagCloud[tag])
    }
	tmpl := template.HTML(tagCloudHTML)
	generatedTagCloudMap[idx] = &tmpl
	return *generatedTagCloudMap[idx]
}
