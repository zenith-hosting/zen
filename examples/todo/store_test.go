package main

import "testing"

func TestTodoStoreAddListDelete(t *testing.T) {
	store := NewTodoStore()

	store.Add("write example")
	store.Add("delete it")

	todos := store.List()
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	if todos[0].ID == 0 || todos[0].Text != "write example" {
		t.Fatalf("unexpected first todo: %#v", todos[0])
	}

	if !store.Delete(todos[0].ID) {
		t.Fatalf("expected delete to succeed")
	}

	todos = store.List()
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo after delete, got %d", len(todos))
	}
	if todos[0].Text != "delete it" {
		t.Fatalf("unexpected remaining todo: %#v", todos[0])
	}
}
