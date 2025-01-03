package helpers

import (
    "html/template"
    "net/url"
    "strconv"

    "github.com/sokkalf/hubro/index"
)

// Paginator preserves existing query parameters by parsing them
func Paginator(currentURL *url.URL, page, totalPages int, entries []index.IndexEntry) template.HTML {
    paginator := `<nav class="pagination" aria-label="pagination">`
    if totalPages > 1 {
        paginator += `<ul class="pagination">`

        // Determine previous page link
        if page > 1 {
            prevPage := page - 1
            paginator += buildPageLink(currentURL, prevPage, "Previous")
        }

        // Links for each page
        for i := 1; i <= totalPages; i++ {
            if i == page {
                // Mark current page
                paginator += buildPageLink(currentURL, i, strconv.Itoa(i), true)
            } else {
                paginator += buildPageLink(currentURL, i, strconv.Itoa(i))
            }
        }

        // Determine next page link
        if page < totalPages {
            nextPage := page + 1
            paginator += buildPageLink(currentURL, nextPage, "Next")
        }

        paginator += "</ul>"
    }
    paginator += "</nav>"

    return template.HTML(paginator)
}

// buildPageLink sets the `p` query parameter for a page, preserves others, and returns the link HTML
func buildPageLink(u *url.URL, page int, text string, isCurrent ...bool) string {
    // Copy the URL so we don’t modify the original pointer
    newURL := *u

    q := newURL.Query()
    q.Set("p", strconv.Itoa(page))   // set/update the "p" param
    newURL.RawQuery = q.Encode()

    cssClasses := "pagination"
    if len(isCurrent) > 0 && isCurrent[0] {
        cssClasses += " is-current"
    }

    return `<li><a data-hx-boost="true" href="` + newURL.String() + `" class="` + cssClasses + `">` + text + `</a></li>`
}
