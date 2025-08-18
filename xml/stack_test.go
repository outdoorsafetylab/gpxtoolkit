package xml

import (
	"testing"
)

func TestStack_Push(t *testing.T) {
	s := &Stack{}

	// Test pushing elements
	s.Push("first")
	s.Push("second")
	s.Push("third")

	if len(s.slice) != 3 {
		t.Errorf("Expected stack length 3, got %d", len(s.slice))
	}

	if s.slice[0] != "first" || s.slice[1] != "second" || s.slice[2] != "third" {
		t.Errorf("Expected [first second third], got %v", s.slice)
	}
}

func TestStack_Pop(t *testing.T) {
	s := &Stack{}
	s.slice = []string{"first", "second", "third"}

	// Test popping elements
	last := s.Pop()
	if last != "third" {
		t.Errorf("Expected 'third', got '%s'", last)
	}

	if len(s.slice) != 2 {
		t.Errorf("Expected stack length 2, got %d", len(s.slice))
	}

	// Test popping from empty stack
	emptyStack := &Stack{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when popping from empty stack")
		}
	}()
	emptyStack.Pop()
}

func TestStack_Peek(t *testing.T) {
	s := &Stack{}

	// Test peeking empty stack
	if s.Peek() != "" {
		t.Errorf("Expected empty string for empty stack, got '%s'", s.Peek())
	}

	// Test peeking with elements
	s.Push("first")
	if s.Peek() != "first" {
		t.Errorf("Expected 'first', got '%s'", s.Peek())
	}

	s.Push("second")
	if s.Peek() != "second" {
		t.Errorf("Expected 'second', got '%s'", s.Peek())
	}
}

func TestStack_Depth(t *testing.T) {
	s := &Stack{}

	// Test empty stack
	if s.Depth() != 0 {
		t.Errorf("Expected depth 0, got %d", s.Depth())
	}

	// Test with elements
	s.Push("first")
	if s.Depth() != 1 {
		t.Errorf("Expected depth 1, got %d", s.Depth())
	}

	s.Push("second")
	if s.Depth() != 2 {
		t.Errorf("Expected depth 2, got %d", s.Depth())
	}

	s.Pop()
	if s.Depth() != 1 {
		t.Errorf("Expected depth 1 after pop, got %d", s.Depth())
	}
}

func TestStack_Clone(t *testing.T) {
	original := &Stack{}
	original.slice = []string{"first", "second", "third"}

	cloned := original.Clone()

	// Test that clone has same elements
	if len(cloned.slice) != len(original.slice) {
		t.Errorf("Expected cloned length %d, got %d", len(original.slice), len(cloned.slice))
	}

	for i, elem := range original.slice {
		if cloned.slice[i] != elem {
			t.Errorf("Expected cloned[%d] = '%s', got '%s'", i, elem, cloned.slice[i])
		}
	}

	// Test that clone is independent
	cloned.Push("fourth")
	if len(original.slice) != 3 {
		t.Errorf("Original stack should not be affected, expected length 3, got %d", len(original.slice))
	}

	original.Push("fourth")
	if len(cloned.slice) != 4 {
		t.Errorf("Cloned stack should not be affected by original, expected length 4, got %d", len(cloned.slice))
	}
}

func TestStack_Contains(t *testing.T) {
	s1 := &Stack{}
	s1.slice = []string{"root", "parent", "child"}

	s2 := &Stack{}
	s2.slice = []string{"root", "parent"}

	s3 := &Stack{}
	s3.slice = []string{"root", "parent", "child", "grandchild"}

	// Test contains with smaller stack
	if !s1.Contains(s2) {
		t.Error("Expected s1 to contain s2")
	}

	// Test contains with larger stack
	if s1.Contains(s3) {
		t.Error("Expected s1 to not contain s3")
	}

	// Test contains with same stack
	if !s1.Contains(s1) {
		t.Error("Expected s1 to contain itself")
	}

	// Test contains with empty stack
	emptyStack := &Stack{}
	if !s1.Contains(emptyStack) {
		t.Error("Expected s1 to contain empty stack")
	}

	// Test contains with different elements
	s4 := &Stack{}
	s4.slice = []string{"root", "different"}
	if s1.Contains(s4) {
		t.Error("Expected s1 to not contain s4 with different elements")
	}
}
