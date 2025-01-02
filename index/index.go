package index

import (
	"fmt"
	"html/template"
	"log/slog"
	"slices"
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
	FileName    string                 `json:"fileName"`
}

const (
	SortBySortOrder = iota
	SortByDate
)

type Index struct {
	Entries       []IndexEntry `json:"entries"`
	ResetChan     chan bool
	FeedResetChan chan bool
	rootPath      string
	name          string
	lookup        map[string]*IndexEntry
	lookupMutex   sync.RWMutex
	sortMutex     sync.Mutex
	sortMode      int
}

func NewIndex(name string, rootPath string) *Index {
	if indices == nil {
		indices = make(Indices)
	}

	if i, ok := indices[name]; ok {
		slog.Error("Index already exists", "name", name)
		return i
	}

	entry := &Index{name: name,
		rootPath:      rootPath,
		lookup:        make(map[string]*IndexEntry),
		sortMode:      SortBySortOrder,
		ResetChan:     make(chan bool),
		FeedResetChan: make(chan bool),
	}
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

func (i *Index) SetSortMode(mode int) {
	if mode == SortBySortOrder || mode == SortByDate {
		i.sortMode = mode
	} else {
		slog.Warn("Invalid sort mode", "mode", mode)
		i.sortMode = SortBySortOrder
	}
}

func (i *Index) AddEntry(e IndexEntry) error {
	if e.Id == "" {
		return fmt.Errorf("entry ID cannot be empty")
	}
	if i.GetEntry(e.Id) != nil {
		return fmt.Errorf("entry with ID %s already exists", e.Id)
	}
	e.Path = i.rootPath + e.Path
	i.lookupMutex.Lock()
	i.Entries = append(i.Entries, e)
	i.lookup[e.Id] = &e
	i.lookupMutex.Unlock()
	return nil
}

func (i *Index) UpdateEntry(e IndexEntry) error {
	if e.Id == "" {
		return fmt.Errorf("entry ID cannot be empty")
	}
	if i.GetEntry(e.Id) == nil {
		return fmt.Errorf("entry with ID %s does not exist", e.Id)
	}
	e.Path = i.rootPath + e.Path
	i.lookupMutex.Lock()
	for j, entry := range i.Entries {
		if entry.Id == e.Id {
			i.Entries[j] = e
			i.lookup[e.Id] = &e
			break
		}
	}
	i.lookupMutex.Unlock()
	return nil
}

func (i *Index) DeleteEntry(id string) error {
	if i.GetEntry(id) == nil {
		return fmt.Errorf("entry with ID %s does not exist", id)
	}
	i.lookupMutex.Lock()
	for j, entry := range i.Entries {
		if entry.Id == id {
			slog.Info("Deleting entry", "id", id)
			i.Entries = slices.Delete(i.Entries, j, j+1)
			delete(i.lookup, id)
			break
		}
	}
	i.lookupMutex.Unlock()
	return nil
}

func (i *Index) DeleteEntryByFileName(path string) error {
	var entry *IndexEntry
	for idx := range i.Entries {
		if i.Entries[idx].FileName == path {
			entry = &i.Entries[idx]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("entry with path %s does not exist", path)
	}
	return i.DeleteEntry(entry.Id)
}

func (i *Index) GetEntry(id string) *IndexEntry {
	i.lookupMutex.RLock()
	defer i.lookupMutex.RUnlock()
	return i.lookup[id]
}

func (i *Index) Sort() {
	switch i.sortMode {
	case SortBySortOrder:
		i.SortBySortOrder()
	case SortByDate:
		i.SortByDate()
	default:
		slog.Warn("Invalid sort mode", "mode", i.sortMode)
	}
}

func (i *Index) SortBySortOrder() {
	i.sortMutex.Lock()
	if i.Entries == nil {
		return
	}
	sort.Slice(i.Entries, func(j, k int) bool {
		return i.Entries[j].SortOrder < i.Entries[k].SortOrder
	})
	i.sortMutex.Unlock()
}

func (i *Index) SortByDate() {
	i.sortMutex.Lock()
	if i.Entries == nil {
		return
	}
	sort.Slice(i.Entries, func(j, k int) bool {
		return i.Entries[k].Date.Before(i.Entries[j].Date)
	})
	i.sortMutex.Unlock()
}

func (i *Index) Count() int {
	return len(i.Entries)
}
