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

func TestUpdateEntry(t *testing.T) {
	name := "updateEntryTest"
	rootPath := "/content/"
	delete(indices, name) // ensure a clean slate before we start

	idx := NewIndex(name, rootPath)

	t.Run("empty ID returns error", func(t *testing.T) {
		err := idx.UpdateEntry(IndexEntry{
			Id:    "", // intentionally empty
			Title: "No ID Entry",
			Path:  "no-id.html",
		})
		if err == nil {
			t.Errorf("expected error when updating entry with empty ID, got nil")
		}
	})

	t.Run("non-existent ID returns error", func(t *testing.T) {
		err := idx.UpdateEntry(IndexEntry{
			Id:    "doesNotExist",
			Title: "Non-existent Entry",
			Path:  "not-found.html",
		})
		if err == nil {
			t.Errorf("expected error when updating non-existent entry, got nil")
		}
	})

	t.Run("valid ID updates successfully", func(t *testing.T) {
		// Add an entry first
		entry := IndexEntry{
			Id:    "validID",
			Title: "Original Title",
			Path:  "original.html",
		}
		idx.AddEntry(entry)

		// Update it
		updated := IndexEntry{
			Id:    "validID",
			Title: "Updated Title",
			Path:  "updated.html",
		}
		err := idx.UpdateEntry(updated)
		if err != nil {
			t.Fatalf("unexpected error updating entry: %v", err)
		}

		// Make sure the new data is in place
		got := idx.GetEntry("validID")
		if got == nil {
			t.Fatalf("expected to retrieve updated entry with ID validID, got nil")
		}
		if got.Title != "Updated Title" {
			t.Errorf("expected updated title 'Updated Title', got '%s'", got.Title)
		}
		// Check that the Path was updated and prefixed with rootPath
		expectedPath := "/content/updated.html"
		if got.Path != expectedPath {
			t.Errorf("expected path '%s', got '%s'", expectedPath, got.Path)
		}
	})
}

// TestDeleteEntry verifies deleting an entry by ID.
func TestDeleteEntry(t *testing.T) {
    name := "deleteEntryTest"
    rootPath := "/content/"
    delete(indices, name) // clean slate

    idx := NewIndex(name, rootPath)

    t.Run("non-existent ID returns error", func(t *testing.T) {
        err := idx.DeleteEntry("doesNotExist")
        if err == nil {
            t.Errorf("expected error for deleting non-existent ID, got nil")
        }
    })

    t.Run("valid ID deletes successfully", func(t *testing.T) {
        entry := IndexEntry{
            Id:       "entry-123",
            FileName: "my-file.md",
            Title:    "Some Title",
            Path:     "my-file.html",
        }
        idx.AddEntry(entry)

        if idx.Count() != 1 {
            t.Fatalf("expected 1 entry before delete, got %d", idx.Count())
        }

        err := idx.DeleteEntry("entry-123")
        if err != nil {
            t.Fatalf("unexpected error deleting entry: %v", err)
        }

        if idx.Count() != 0 {
            t.Errorf("expected 0 entries after delete, got %d", idx.Count())
        }
        if idx.GetEntry("entry-123") != nil {
            t.Errorf("expected nil when retrieving deleted entry, got non-nil")
        }
    })
}
