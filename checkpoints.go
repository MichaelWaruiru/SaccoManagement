package main

import (
	"html/template"
	"log"
	"net/http"
)

type Checkpoint struct {
	ID             int
	CheckpointName string
	SaccoName      string
	IsSelected     bool
}

func checkpointsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	session, _ := store.Get(r, "sacco-mgmnt")
	if session.Values["user"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	checkpoints, err := getAllCheckpoints()
	if err != nil {
		http.Error(w, "failed to get checkpoints", http.StatusInternalServerError)
		return
	}

	data := struct {
		Checkpoints []Checkpoint
	}{
		Checkpoints: checkpoints,
	}

	// Execute menu template
	menuTemplate := template.Must(template.ParseFiles("templates/menu.html"))

	// Execute menu template
	err = menuTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Println("Error executing menu template:", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "checkpoints", data)
	if err != nil {
		http.Error(w, "Error rendering checkpoints", http.StatusInternalServerError)
		return
	}
}

func getAllCheckpoints() ([]Checkpoint, error) {
	var checkpoints []Checkpoint

	rows, err := db.Query("SELECT * FROM checkpoints")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var checkpoint Checkpoint
		err := rows.Scan(&checkpoint.ID, &checkpoint.CheckpointName)
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

func addCheckpointsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	checkpointName := r.FormValue("checkpointName")

	if checkpointName == "" {
		http.Error(w, "Missing required form fields", http.StatusBadRequest)
		return
	}

	newCheckpoint := Checkpoint{
		CheckpointName: checkpointName,
	}

	err = addCheckpoint(newCheckpoint)
	if err != nil {
		http.Error(w, "Failed to add checkpoint", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/checkpoints", http.StatusSeeOther)
}

func addCheckpoint(checkpoint Checkpoint) error {
	_, err := db.Exec("INSERT INTO checkpoints (checkpoint_name) VALUES (?)", checkpoint.CheckpointName)
	if err != nil {
		return err
	}

	return nil
}
