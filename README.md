_ = s.Set("key1", "value1")
	if s.Size() != 1 {
		t.Errorf("After one insert Size() = %d, want 1", s.Size())
	}

	_ = s.Set("key2", "value2")
	if s.Size() != 2 {
		t.Errorf("After two inserts Size() = %d, want 2", s.Size())
	}

	_ = s.Delete("key1")
	if s.Size() != 1 {
		t.Errorf("After one delete Size() = %d, want 1", s.Size())
	}
}

func TestClear(t *testing.T) {
	s := New()
	_ = s.Set("key1", "value1")
	_ = s.Set("key2", "value2")

	s.Clear()

	if s.Size() != 0 {
		t.Errorf("After Clear() Size() = %d, want 0", s.Size())
	}

	if s.Exists("key1") || s.Exists("key2") {
		t.Error("Keys still exist after Clear()")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	// Number of concurrent goroutines
	numGoroutines := 100
	numOperations := 1000

	wg.Add(numGoroutines * 3) // Set, Get, Delete operations

	// Concurrent Set operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key"
				value := "value"
				_ = s.Set(key, value)
			}
		}(i)
	}

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key"
				_, _ = s.Get(key)
			}
		}(i)
	}

	// Concurrent Delete operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key"
				_ = s.Delete(key)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions occur
}
