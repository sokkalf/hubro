package index

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestNewIndex checks if a new index can be created and stored in the global indices.
func TestNewIndex(t *testing.T) {
	name := "testIndex"
	rootPath := "/test/"

	// Clean up any prior leftover
	delete(indices, name)

	idx := NewIndex(name, rootPath)
	if idx == nil {
		t.Fatalf("Expected new Index to be created, got nil")
	}
	if idx.GetName() != name {
		t.Errorf("Index name mismatch. Want %s, got %s", name, idx.GetName())
	}
	if idx.rootPath != rootPath {
		t.Errorf("Index rootPath mismatch. Want %s, got %s", rootPath, idx.rootPath)
	}

	// Creating again with the same name should log an error and return the same index
	idx2 := NewIndex(name, rootPath)
	if idx2 != idx {
		t.Errorf("Expected same index to be returned when creating an existing index name")
	}
}

// TestAddEntry ensures entries can be added and retrieved.
func TestAddEntry(t *testing.T) {
	name := "addEntryIndex"
	rootPath := "/content/"
	delete(indices, name) // reset global map before test

	idx := NewIndex(name, rootPath)
	entry := IndexEntry{
		Id:    "testID",
		Title: "Test Title",
		Path:  "page1.html",
	}

	idx.AddEntry(entry)

	if len(idx.Entries) != 1 {
		t.Fatalf("Expected 1 entry in index, got %d", len(idx.Entries))
	}

	// Check that the path was prefixed with rootPath
	if idx.Entries[0].Path != "/content/page1.html" {
		t.Errorf("Expected path to be /content/page1.html, got %s", idx.Entries[0].Path)
	}

	// Check lookup
	got := idx.GetEntry("testID")
	if got == nil {
		t.Fatalf("Expected to get entry from lookup, got nil")
	}
	if got.Title != "Test Title" {
		t.Errorf("Expected Title = 'Test Title', got '%s'", got.Title)
	}
}

// TestGetIndex ensures we can retrieve an index from the global indices map.
func TestGetIndex(t *testing.T) {
	name := "getIndexTest"
	rootPath := "/"
	delete(indices, name)

	NewIndex(name, rootPath)
	idx := GetIndex(name)
	if idx == nil {
		t.Fatalf("Expected to retrieve an index named %q, got nil", name)
	}

	if GetIndex("nonexistent") != nil {
		t.Errorf("Expected nil for nonexisting index, got non-nil")
	}
}

// TestSortBySortOrder checks sorting by SortOrder.
func TestSortBySortOrder(t *testing.T) {
	name := "sortOrderIndex"
	rootPath := "/"
	delete(indices, name)

	idx := NewIndex(name, rootPath)

	idx.AddEntry(IndexEntry{Id: "3", SortOrder: 3})
	idx.AddEntry(IndexEntry{Id: "1", SortOrder: 1})
	idx.AddEntry(IndexEntry{Id: "2", SortOrder: 2})

	idx.SortBySortOrder()

	ids := []string{idx.Entries[0].Id, idx.Entries[1].Id, idx.Entries[2].Id}
	expected := []string{"1", "2", "3"}
	for i, want := range expected {
		if ids[i] != want {
			t.Errorf("SortBySortOrder mismatch at pos %d: want %s, got %s", i, want, ids[i])
		}
	}
}

// TestSortByDate checks sorting by date (descending).
func TestSortByDate(t *testing.T) {
	name := "sortDateIndex"
	rootPath := "/"
	delete(indices, name)

	idx := NewIndex(name, rootPath)
	now := time.Now()

	idx.AddEntry(IndexEntry{Id: "oldest", Date: now.Add(-48 * time.Hour)})
	idx.AddEntry(IndexEntry{Id: "newest", Date: now.Add(48 * time.Hour)})
	idx.AddEntry(IndexEntry{Id: "middle", Date: now})

	idx.SortByDate()
	// SortByDate sorts so that the newest (largest time) appears first
	// i.e., i.Entries[0] has the largest Date
	//   i.Entries[last] has the smallest Date

	if idx.Entries[0].Id != "newest" {
		t.Errorf("Expected 'newest' entry first, got %s", idx.Entries[0].Id)
	}
	if idx.Entries[1].Id != "middle" {
		t.Errorf("Expected 'middle' entry second, got %s", idx.Entries[1].Id)
	}
	if idx.Entries[2].Id != "oldest" {
		t.Errorf("Expected 'oldest' entry third, got %s", idx.Entries[2].Id)
	}
}

// TestConcurrentAdd tests adding entries from multiple goroutines.
func TestConcurrentAdd(t *testing.T) {
	name := "concurrentIndex"
	rootPath := "/"
	delete(indices, name)

	idx := NewIndex(name, rootPath)

	const numGoroutines = 10
	const entriesPerGoroutine = 100
	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for e := 0; e < entriesPerGoroutine; e++ {
				entry := IndexEntry{
					Id:    fmt.Sprintf("entry-%d-%d", gid, e),
					Title: "Some Title",
					Path:  "page.html",
				}
				idx.AddEntry(entry)
			}
		}(g)
	}

	wg.Wait()

	totalEntries := numGoroutines * entriesPerGoroutine
	if idx.Count() != totalEntries {
		t.Errorf("Expected %d entries after concurrency test, got %d", totalEntries, idx.Count())
	}
}
