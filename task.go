package main

import (
	"fmt"
	"os"

	"time"

	"gopkg.in/go-playground/validator.v9"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var validate *validator.Validate

func getDatabase() (*mgo.Session, *mgo.Collection, error) {
	// Connection to mongodb server.
	host := "localhost"
	session, err := mgo.Dial(host)
	if err != nil {
		return nil, nil, fmt.Errorf("can't to connect to mongodb server at %v (%v)", host, err)
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	// Select the collection.
	c := session.DB(os.Getenv("TASK_DB")).C("tasks")

	return session, c, nil
}

// Task is the type of a task.
type Task struct {
	ID        bson.ObjectId `bson:"_id,omitempty" `
	SID       string        `bson:"sid,omitempty" jsonapi:"primary,task"`
	Title     string        `bson:"title" validate:"required" jsonapi:"attr,title"`
	Done      bool          `bson:"done" jsonapi:"attr,done"`
	CreatedAt time.Time     `bson:"createdAt" jsonapi:"attr,created_at"`
	UpdatedAt time.Time     `bson:"updatedAt" jsonsapi:"attr,updated_at"`
}

// NewTask create a new task.
func NewTask(title string) (*Task, error) {
	task := &Task{Title: title, CreatedAt: time.Now(), Done: false}
	if err := task.Validate(); err != nil {
		return nil, err
	}
	return task, nil
}

// Validate checks attributes's integrity.
func (t *Task) Validate() error {
	//errs := validator.Validate(t)
	validate = validator.New()
	err := validate.Struct(t)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			panic(err)
		}
		return err
	}
	return nil
}

// SelectTask find a task by ID or Title.
func SelectTask(query string) (*Task, error) {

	// Get the DB.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return nil, err
	}

	// Check query parameter.
	if query == "" {
		return nil, fmt.Errorf("query parameter is empty %v", query)
	}

	// Find the task.
	t := &Task{}
	var q *mgo.Query
	if bson.IsObjectIdHex(query) {
		q = c.Find(bson.M{"_id": bson.ObjectIdHex(query)})
	} else {
		q = c.Find(bson.M{"title": query})
	}

	// Check the count and return an empty task.
	if n, _ := q.Count(); n == 0 {
		return t, nil
	}

	// Get the task.
	if err = q.One(t); err != nil {
		return nil, err
	}

	return t, nil
}

// SearchTask find all tasks with parameters.
func SearchTask(query string, done bool, all bool, page int, limit int) ([]*Task, int, error) {

	// Get the DB.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return nil, 0, err
	}
	reg := bson.RegEx{Pattern: query, Options: ""}
	bq := bson.M{"title": reg}

	if !all {
		bq["done"] = done
	}

	q := c.Find(bq)

	n, err := q.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("unexpected error %v", err)
	}

	q.Sort("title", "_id").Limit(limit)

	// To get the nth page:
	q = q.Skip((page - 1) * limit)

	var tasks []*Task
	if err = q.All(&tasks); err != nil {
		return nil, 0, fmt.Errorf("unexpected error %v", err)
	}

	return tasks, n, nil
}

// Save persist the task into the database.
func (t *Task) Save() error {
	// Get the database connection.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return err
	}

	// Find an existing task with the same properties.
	r, err := SelectTask(t.Title)
	if err != nil {
		return err
	}
	if (Task{}) != *r {
		return fmt.Errorf("task already exists %v", r.SID)
	}

	// Generete a mongoDB and Json ID.
	t.ID = bson.NewObjectId()
	t.SID = t.ID.Hex()

	t.CreatedAt = time.Now()

	// Persist the task.
	err = c.Insert(&t)
	if err != nil {
		return fmt.Errorf("can't to persist the task (%v)", err)
	}

	return nil
}

// Update persist an existing task with new properties
func (t *Task) Update() error {
	// Get the database connection.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return err
	}

	// Check the ID.
	if !t.ID.Valid() {
		if bson.IsObjectIdHex(t.SID) {
			t.ID = bson.ObjectIdHex(t.SID)
		} else {
			return fmt.Errorf("ID is required for update task")
		}
	}

	// Persist the task.
	if err := c.UpdateId(t.ID, bson.M{"$set": bson.M{"title": t.Title, "done": t.Done, "updatedAt": time.Now()}}); err != nil {
		return fmt.Errorf("can't to persist the task (%v)", err)
	}

	return nil
}

// Delete remove an existing task.
func (t *Task) Delete() error {
	// Get the database connection.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return err
	}

	// Remove the task
	err = c.Remove(bson.M{"_id": t.ID})
	if err != nil {
		return err
	}

	return nil
}

// DeleteTask remove a task.
func DeleteTask(id string) error {
	// Get the database connection.
	s, c, err := getDatabase()
	defer s.Close()
	if err != nil {
		return err
	}

	// Remove the task
	if !bson.IsObjectIdHex(id) {
		return fmt.Errorf("id value is not valid (%v)", id)
	}
	oid := bson.ObjectIdHex(id)
	err = c.Remove(bson.M{"_id": oid})
	if err != nil {
		return err
	}

	return nil
}
