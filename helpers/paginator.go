package helpers

import (
	"html/template"
	"strconv"

	"github.com/sokkalf/hubro/index"
)

func Paginator(page int, totalPages int, entries []index.IndexEntry) template.HTML {
	var prevPage, nextPage int
	if page > 1 {
		prevPage = page - 1
	}
	if page < totalPages {
		nextPage = page + 1
	}

	paginator := `<nav class="pagination" role="navigation" aria-label="pagination">`
	if totalPages > 1 {
		paginator += `<ul class="pagination">`
		if prevPage > 0 {
			paginator += `<li><a hx-boost="true" href="?p=` + strconv.Itoa(prevPage) + `" class="pagination">Previous</a></li>`
		}
		for i := 1; i <= totalPages; i++ {
			if i == page {
				paginator += `<li><a hx-boost="true" href="?p=` + strconv.Itoa(i) + `" class="pagination is-current">` + strconv.Itoa(i) + `</a></li>`
			} else {
				paginator += `<li><a hx-boost="true" href="?p=` + strconv.Itoa(i) + `" class="pagination">` + strconv.Itoa(i) + `</a></li>`
			}
		}
		if nextPage > 0 {
			paginator += `<li><a hx-boost="true" href="?p=` + strconv.Itoa(nextPage) + `" class="pagination">Next</a></li>`
		}
		paginator += "</ul>"
	}
	paginator += "</nav>"
	return template.HTML(paginator)
}
