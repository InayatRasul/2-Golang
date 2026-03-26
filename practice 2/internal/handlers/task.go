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

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		idStr := r.URL.Query().Get("id")
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
				return
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