package helpers

import (
	"fmt"
	"html/template"

	"github.com/sokkalf/hubro/config"
	"github.com/sokkalf/hubro/index"
)

var generatedTagCloud *template.HTML

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
	if generatedTagCloud != nil {
		return *generatedTagCloud
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

	var tagCloudHTML string
	tagHTML := func(tag string, count int) string {
		return fmt.Sprintf(`<span class="%s"><a href="%s?tag=%s">%s</a></span>%s`,
			cssTextSize(count), config.Config.RootPath, tag, tag, "\n")
	}
	for tag, count := range tagCloud {
		tagCloudHTML += tagHTML(tag, count)
	}
	tmpl := template.HTML(tagCloudHTML)
	generatedTagCloud = &tmpl
	return *generatedTagCloud
}
