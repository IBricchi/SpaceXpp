/*
   Written by Bradley Stanley-Clamp (bradley.stanley-clamp19@imperial.ac.uk) and Nicholas Pfaff (nicholas.pfaff19@imperial.ac.uk), 2021 - SpaceX++ EEE/EIE 2nd year group project, Imperial College London
*/

package server

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type staticTestData struct {
	Info string `json:"info"`
	Data []int  `json:"data"`
}

type rover struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Rotation int `json:"rotation"`
}

type energy struct {
	StateOfCharge int `json:"stateOfCharge"`
	StateOfHealth int `json:"stateOfHealth"`
	ErrorInCells  int `json:"errorInCells"`
}

func (h *HttpServer) connect(w http.ResponseWriter, req *http.Request) {
	status := h.mqtt.getIsConnected()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

}

func (h *HttpServer) battery(w http.ResponseWriter, req *http.Request) {

}

func check(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	data := staticTestData{
		Info: "Some static test data",
		Data: []int{
			1, 2, 3, 4, 5,
		},
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HttpServer) updateWebMap(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	data := Map
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HttpServer) updateRover(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	data := Rover
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HttpServer) loadMap(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		data := dbMap
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// clearning instruction log
		var empty []driveInstruction
		dbMap.Instructions = empty

		w.WriteHeader(http.StatusOK)
	}
}

func (h *HttpServer) getFeed(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		data := feed
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// clearing feed
		var empty string
		feed = empty

		w.WriteHeader(http.StatusOK)
	}
}

func (h *HttpServer) getEnergyStatus(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(currentEnergy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

}

func (h *HttpServer) getIsAuthorised(creds map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		data := true

		username, password, ok := r.BasicAuth()
		if !ok {
			data = false
		}

		passwordHash, userExists := creds[username]
		if !userExists {
			data = false
		}

		passwordHashBytes := []byte(passwordHash)
		passwordBytes := []byte(password)
		if err := bcrypt.CompareHashAndPassword(passwordHashBytes, passwordBytes); err != nil {
			data = false
		}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
