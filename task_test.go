package main

import (
	"reflect"
	"testing"
)

func newTaskOrFatal(t *testing.T, title string) *Task {
	task, err := NewTask(title)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	return task
}

func createTaskOrFatal(t *testing.T, title string) *Task {
	task, err := NewTask(title)
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if err := task.Save(); err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	return task
}

func selectTaskOrFatal(t *testing.T, title string) *Task {
	task, err := SelectTask(title)
	if err != nil {
		t.Fatalf("unexpected error (%v)", err)
	}
	return task
}

func TestNewTask(t *testing.T) {
	title := "testing task"
	task := newTaskOrFatal(t, title)
	if task.Title != title {
		t.Errorf("expected title '%v' and got '%v'", title, task.Title)
	}
}

func TestNewTaskWithEmptyTitle(t *testing.T) {
	_, err := NewTask("")
	if err == nil {
		t.Errorf("expected error 'empty title', got %v", err)
	}
}

func TestSaveTask(t *testing.T) {
	task := newTaskOrFatal(t, "testing task")
	err := task.Save()
	if err != nil {
		t.Errorf("unexpected error : %v", err)
	}
}

func TestFindTaskByTitle(t *testing.T) {
	task := createTaskOrFatal(t, "test select task by title")
	find, err := SelectTask(task.Title)
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	if task.Title != find.Title {
		t.Errorf("expected task with title '%v', got '%v'", task.Title, find.Title)
	}
}

func TestSelectTaskByID(t *testing.T) {
	task := createTaskOrFatal(t, "test select task by id")
	find, err := SelectTask(task.SID)
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	if find.SID != task.SID {
		t.Errorf("expected task id %v, got %v", task.SID, find.SID)
	}
}

func TestSearchTask(t *testing.T) {

	// Check search by title.
	tasks, n, err := SearchTask("search", false, true, 1, 10)
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	// Check the count.
	if n != 100 {
		t.Errorf("expected count 100, got %v", n)
	}

	// Check the limit.
	if len(tasks) != 10 {
		t.Errorf("expected limit 10, got %v", len(tasks))
	}

	// Check the pagination.
	tasks2, n, err := SearchTask("search", false, true, 2, 10)
	if reflect.DeepEqual(tasks, tasks2) {
		t.Errorf("page 1 is not different to page 2")
	}

	// Check the done task.
	tasks, n, err = SearchTask("search", true, false, 2, 10)
	if n != 50 {
		t.Errorf("expected 50 done task, got %v", n)
	}

	// Check the not done task.
	tasks, n, err = SearchTask("search", false, false, 2, 10)
	if n != 50 {
		t.Errorf("expected 50 not done task, got %v", n)
	}

	// Test empty query
	tasks, n, err = SearchTask("", false, false, 2, 10)
	if n == 0 {
		t.Errorf("expected more than 0, got %v", n)
	}
}

func TestSaveNewExistingTask(t *testing.T) {
	task := newTaskOrFatal(t, "testing task")
	err := task.Save()
	if err == nil {
		t.Errorf("expected error (%v)", err)
	}
}

func TestUpdateTask(t *testing.T) {
	// Create a new task.
	task := createTaskOrFatal(t, "test task update task")

	// Update the title.
	title := "test task with an updated title"
	task.Title = title
	err := task.Update()
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	// Find the task by new title.
	task2 := selectTaskOrFatal(t, title)
	if task2 == nil {
		t.Errorf("expected an updated task, got %#v", t)
	}
}

func TestUpdateNewTask(t *testing.T) {
	task := newTaskOrFatal(t, "testing task")
	err := task.Update()
	if err == nil {
		t.Errorf("expected an error, got %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	title := "test delete task"
	task := createTaskOrFatal(t, title)

	if err := DeleteTask(task.SID); err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	dt := selectTaskOrFatal(t, title)
	if (Task{}) != *dt {
		t.Errorf("expected an empty task, got %v", dt)
	}
}

func TestDeleteNewTask(t *testing.T) {
	task := newTaskOrFatal(t, "test delete new task")
	if err := DeleteTask(task.SID); err == nil {
		t.Errorf("expected an error, got %v", err)
	}
}
