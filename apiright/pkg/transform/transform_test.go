package transform

import (
	"testing"
)

// Test models
type SourceModel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type TargetModel struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
}

func TestTransformer_Transform(t *testing.T) {
	source := SourceModel{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	transformer := NewTransformer[SourceModel, TargetModel]().
		WithFieldMapping("Name", "FullName")

	target, err := transformer.Transform(source)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if target.ID != source.ID {
		t.Errorf("Expected ID %d, got %d", source.ID, target.ID)
	}

	if target.FullName != source.Name {
		t.Errorf("Expected FullName %s, got %s", source.Name, target.FullName)
	}

	if target.Email != source.Email {
		t.Errorf("Expected Email %s, got %s", source.Email, target.Email)
	}

	if target.Age != source.Age {
		t.Errorf("Expected Age %d, got %d", source.Age, target.Age)
	}
}

func TestTransformer_TransformSlice(t *testing.T) {
	sources := []SourceModel{
		{ID: 1, Name: "John", Email: "john@example.com", Age: 30},
		{ID: 2, Name: "Jane", Email: "jane@example.com", Age: 25},
	}

	transformer := NewTransformer[SourceModel, TargetModel]().
		WithFieldMapping("Name", "FullName")

	targets, err := transformer.TransformSlice(sources)
	if err != nil {
		t.Fatalf("TransformSlice failed: %v", err)
	}

	if len(targets) != len(sources) {
		t.Errorf("Expected %d targets, got %d", len(sources), len(targets))
	}

	for i, target := range targets {
		source := sources[i]
		if target.ID != source.ID {
			t.Errorf("Target %d: Expected ID %d, got %d", i, source.ID, target.ID)
		}
		if target.FullName != source.Name {
			t.Errorf("Target %d: Expected FullName %s, got %s", i, source.Name, target.FullName)
		}
	}
}

func TestBiDirectionalTransformer(t *testing.T) {
	source := SourceModel{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	transformer := NewBiDirectionalTransformer[SourceModel, TargetModel]().
		WithFieldMapping("Name", "FullName")

	// Forward transformation
	target, err := transformer.Forward(source)
	if err != nil {
		t.Fatalf("Forward transform failed: %v", err)
	}

	if target.FullName != source.Name {
		t.Errorf("Expected FullName %s, got %s", source.Name, target.FullName)
	}

	// Backward transformation
	backToSource, err := transformer.Backward(target)
	if err != nil {
		t.Fatalf("Backward transform failed: %v", err)
	}

	if backToSource.Name != target.FullName {
		t.Errorf("Expected Name %s, got %s", target.FullName, backToSource.Name)
	}
}

func TestQuickTransform(t *testing.T) {
	source := SourceModel{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	// Test the quick transform function
	target, err := Transform[SourceModel, TargetModel](source)
	if err != nil {
		t.Fatalf("Quick transform failed: %v", err)
	}

	// Should match fields with same names
	if target.ID != source.ID {
		t.Errorf("Expected ID %d, got %d", source.ID, target.ID)
	}

	if target.Email != source.Email {
		t.Errorf("Expected Email %s, got %s", source.Email, target.Email)
	}

	if target.Age != source.Age {
		t.Errorf("Expected Age %d, got %d", source.Age, target.Age)
	}
}