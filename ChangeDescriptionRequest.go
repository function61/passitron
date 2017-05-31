package main

import (
	"encoding/json"
	"net/http"
)

type ChangeDescriptionRequest struct {
	Id          string
	Description string
}

func HandleChangeDescriptionRequest(w http.ResponseWriter, r *http.Request) {
	var req ChangeDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if secretById(req.Id) == nil {
		http.Error(w, "Invalid secret Id", http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		DescriptionChanged{
			Id:          req.Id,
			Description: req.Description,
		},
	})

	w.Write([]byte("OK"))
}
