package helpers

import (
	"fmt"
	"html/template"

	"github.com/sokkalf/hubro/index"
)

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
	tagCloud := tagCloudMap(idx)
	var tagCloudHTML string
	tagHTML := func(tag string, count int) string {
		return " <a href=\"/tags/" + tag + "\" class=\"tag-cloud__tag tag-cloud__tag--" + tag + "\">" + tag + fmt.Sprintf("(%d)", count) + "</a>"
	}
	for tag, count := range tagCloud {
		tagCloudHTML += tagHTML(tag, count)
	}
	return template.HTML(tagCloudHTML)
}
