package server

import (
	"log/slog"
	"time"
)

type Indices map[string]*Index

var indices Indices

type IndexEntry struct {
	Title string `json:"title"`
	Author string `json:"author"`
	Path string `json:"path"`
	Date time.Time `json:"date"`
	SortOrder int `json:"sortOrder"`
	Metadata map[string]interface{} `json:"metadata"`
	Visible bool `json:"visible"`
	HideAuthor bool `json:"hideAuthor"`
	Tags []string `json:"tags"`
}

type Index struct {
	Entries []IndexEntry `json:"entries"`
	rootPath string
	name string
}

func NewIndex(name string, rootPath string) *Index {
	entry := Index{name: name, rootPath: rootPath}
	if indices == nil {
		indices = make(Indices)
		indices[name] = &entry
	} else if _, ok := indices[name]; ok {
		slog.Error("Index already exists", "name", name)
	}
	return &entry
}

func GetIndex(name string) *Index {
	if i, ok := indices[name]; ok {
		return i
	}
	return nil
}

func (i *Index) AddEntry(e IndexEntry) {
	e.Path = i.rootPath + e.Path
	i.Entries = append(i.Entries, e)
}
