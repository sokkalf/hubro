package index

import (
	"fmt"
	"html/template"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/sokkalf/hubro/utils/broker"
)

type Indices map[string]*Index

var indices Indices

type IndexEntry struct {
	Id          string                 `json:"id"`
	Slug        string                 `json:"slug"`
	Title       string                 `json:"title"`
	Author      string                 `json:"author"`
	Path        string                 `json:"path"`
	Date        time.Time              `json:"date"`
	SortOrder   int                    `json:"sortOrder"`
	Metadata    map[string]interface{} `json:"metadata"`
	Visible     bool                   `json:"visible"`
	HideAuthor  bool                   `json:"hideAuthor"`
	HideTitle   bool                   `json:"hideTitle"`
	Tags        []string               `json:"tags"`
	Summary     *template.HTML         `json:"summary"`
	Body        *template.HTML         `json:"body"`
	Description string                 `json:"description"`
	FileName    string                 `json:"fileName"`
}

type Message int

const (
	Updated Message = iota
	Scanned
)

type Index struct {
	entries    []IndexEntry
	rootPath   string
	name       string
	lookup     map[string]*IndexEntry
	slugLookup map[string]*IndexEntry
	mtx        sync.RWMutex
	sortMode   int
	MsgBroker  *broker.Broker[Message]
}

const (
	SortBySortOrder = iota
	SortByDate
)

func NewIndex(name string, rootPath string) *Index {
	if indices == nil {
		indices = make(Indices)
	}

	if i, ok := indices[name]; ok {
		slog.Error("Index already exists", "name", name)
		return i
	}

	entry := &Index{name: name,
		rootPath:   rootPath,
		lookup:     make(map[string]*IndexEntry),
		slugLookup: make(map[string]*IndexEntry),
		sortMode:   SortBySortOrder,
	}

	entry.MsgBroker = broker.NewBroker[Message]()
	go entry.MsgBroker.Start()

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
	i.mtx.Lock()
	defer i.mtx.Unlock()
	i.entries = append(i.entries, e)
	i.lookup[e.Id] = &e
	i.slugLookup[e.Slug] = &e
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
	i.mtx.Lock()
	defer i.mtx.Unlock()
	for j, entry := range i.entries {
		if entry.Id == e.Id {
			i.entries[j] = e
			i.lookup[e.Id] = &e
			i.slugLookup[e.Slug] = &e
			break
		}
	}
	return nil
}

func (i *Index) DeleteEntry(id string) error {
	if i.GetEntry(id) == nil {
		return fmt.Errorf("entry with ID %s does not exist", id)
	}
	i.mtx.Lock()
	defer i.mtx.Unlock()
	for j, entry := range i.entries {
		if entry.Id == id {
			slog.Info("Deleting entry", "id", id)
			i.entries = slices.Delete(i.entries, j, j+1)
			delete(i.lookup, id)
			delete(i.slugLookup, entry.Slug)
			break
		}
	}
	return nil
}

func (i *Index) RLock() {
	i.mtx.RLock()
}

func (i *Index) RUnlock() {
	i.mtx.RUnlock()
}

func (i *Index) Lock() {
	i.mtx.Lock()
}

func (i *Index) Unlock() {
	i.mtx.Unlock()
}

func (i *Index) GetEntries() []IndexEntry {
	i.mtx.RLock()
	defer i.mtx.RUnlock()
	return i.entries
}

func (i *Index) GetEntry(id string) *IndexEntry {
	i.mtx.RLock()
	defer i.mtx.RUnlock()
	return i.lookup[id]
}

func (i *Index) GetEntryBySlug(slug string) *IndexEntry {
	i.mtx.RLock()
	defer i.mtx.RUnlock()
	return i.slugLookup[slug]
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
	i.mtx.Lock()
	defer i.mtx.Unlock()
	if i.entries == nil {
		return
	}
	sort.Slice(i.entries, func(j, k int) bool {
		return i.entries[j].SortOrder < i.entries[k].SortOrder
	})
}

func (i *Index) SortByDate() {
	i.mtx.Lock()
	defer i.mtx.Unlock()
	if i.entries == nil {
		return
	}
	sort.Slice(i.entries, func(j, k int) bool {
		return i.entries[k].Date.Before(i.entries[j].Date)
	})
}

func (i *Index) Count() int {
	i.mtx.RLock()
	defer i.mtx.RUnlock()
	return len(i.entries)
}
