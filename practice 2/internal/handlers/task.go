package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var tasks = []Task{}
var nextID = 1

// there is no variable type at the very end of that line. This means the function returns nothing.
func TaskHandler(w http.ResponseWriter, r *http.Request) {
	// w http.ResponseWriter: This is your Output. 200 404
	// r *http.Request: This is your Input.
	// It contains everything the user sent to you: the URL, the Method (GET/POST), any "Query Parameters" (like ?id=1), and the Request Body.
	w.Header().Set("Content-Type", "application/json")
	// This ensures every response sent by this handler is labeled as JSON.

	switch r.Method {
	// The code checks if the user is trying to "GET" (read) data.
	case http.MethodGet:
		idStr := r.URL.Query().Get("id") //idStr := r.URL.Query().Get("id") looks for "id" in the URL.
		// If it's empty (""), the server simply sends back the entire list of tasks.
		if idStr == "" {
			// GET /tasks - List all
			json.NewEncoder(w).Encode(tasks)
			return
		}

		// GET /tasks?id=X - Find by id
		id, _ := strconv.Atoi(idStr)
		for _, t := range tasks{
			if t.ID == id {
				json.NewEncoder(w).Encode(t)
				//User (the person using Postman or a browser) receives whatever you wrote into w before you hit that return button:
				return
				//"Stop running this function right now and go back to where you came from."
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error":"task not found"})
	
	case http.MethodPost:
		// POST /tasks - create new
		var newTask Task
		if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil || newTask.Title == ""{
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid title"})
			return
		}
		newTask.ID = nextID
		nextID++
		newTask.Done = false
		tasks = append(tasks, newTask)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTask)

	case http.MethodPatch:
		// patch /tasks?id=x - update status
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)

		var updateData struct {
			Done bool `json:"done"`
		}
		json.NewDecoder(r.Body).Decode(&updateData)

		for i, t := range tasks {
			if t.ID == id {
				tasks[i].Done = updateData.Done
				json.NewEncoder(w).Encode(map[string]bool{"updated": true})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error":"task not found"})

	}
}