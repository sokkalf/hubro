package index

import (
	"html/template"
	"log/slog"
	"sort"
	"sync"
	"time"
)

type Indices map[string]*Index

var indices Indices

type IndexEntry struct {
	Id          string                 `json:"id"`
	Title       string                 `json:"title"`
	Author      string                 `json:"author"`
	Path        string                 `json:"path"`
	Date        time.Time              `json:"date"`
	SortOrder   int                    `json:"sortOrder"`
	Metadata    map[string]interface{} `json:"metadata"`
	Visible     bool                   `json:"visible"`
	HideAuthor  bool                   `json:"hideAuthor"`
	Tags        []string               `json:"tags"`
	Summary     *template.HTML         `json:"summary"`
	Body        *template.HTML         `json:"body"`
	Description string                 `json:"description"`
}

type Index struct {
	Entries     []IndexEntry `json:"entries"`
	rootPath    string
	name        string
	lookup      map[string]*IndexEntry
	lookupMutex sync.RWMutex
}

func NewIndex(name string, rootPath string) *Index {
	if indices == nil {
		indices = make(Indices)
	}

	if i, ok := indices[name]; ok {
		slog.Error("Index already exists", "name", name)
		return i
	}

	entry := &Index{name: name, rootPath: rootPath, lookup: make(map[string]*IndexEntry)}
	indices[name] = entry
	return entry
}

func GetIndex(name string) *Index {
	if i, ok := indices[name]; ok {
		return i
	}
	return nil
}

func (i *Index) GetName() string {
	return i.name
}

func (i *Index) AddEntry(e IndexEntry) {
	e.Path = i.rootPath + e.Path
	i.Entries = append(i.Entries, e)
	i.lookupMutex.Lock()
	i.lookup[e.Id] = &e
	i.lookupMutex.Unlock()
}

func (i *Index) GetEntry(id string) *IndexEntry {
	if e, ok := i.lookup[id]; ok {
		return e
	}
	return nil
}

func (i *Index) SortBySortOrder() {
	if i.Entries == nil {
		return
	}
	sort.Slice(i.Entries, func(j, k int) bool {
		return i.Entries[j].SortOrder < i.Entries[k].SortOrder
	})
}

func (i *Index) SortByDate() {
	if i.Entries == nil {
		return
	}
	sort.Slice(i.Entries, func(j, k int) bool {
		return i.Entries[k].Date.Before(i.Entries[j].Date)
	})
}

func (i *Index) Count() int {
	return len(i.Entries)
}
