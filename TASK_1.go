package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Certificate struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

var (
	certificates []Certificate
	mutex        sync.Mutex 
)

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getCertificateByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendJSONResponse(w, map[string]string{"error": "Invalid certificate ID"}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for _, cert := range certificates {
		if cert.ID == id {
			sendJSONResponse(w, cert, http.StatusOK)
			return
		}
	}
	sendJSONResponse(w, map[string]string{"error": "Certificate not found"}, http.StatusNotFound)
}

func createCertificate(w http.ResponseWriter, r *http.Request) {
	var cert Certificate
	if err := json.NewDecoder(r.Body).Decode(&cert); err != nil {
		sendJSONResponse(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	cert.ID = len(certificates) + 1
	certificates = append(certificates, cert)

	sendJSONResponse(w, cert, http.StatusCreated)
}

func getAllCertificates(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	sendJSONResponse(w, certificates, http.StatusOK)
}

func updateCertificate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendJSONResponse(w, map[string]string{"error": "Invalid certificate ID"}, http.StatusBadRequest)
		return
	}

	var updatedCert Certificate
	if err := json.NewDecoder(r.Body).Decode(&updatedCert); err != nil {
		sendJSONResponse(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, cert := range certificates {
		if cert.ID == id {
			updatedCert.ID = id
			certificates[i] = updatedCert
			sendJSONResponse(w, updatedCert, http.StatusOK)
			return
		}
	}
	sendJSONResponse(w, map[string]string{"error": "Certificate not found"}, http.StatusNotFound)
}

func main() {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/certificates/{id}", getCertificateByID).Methods("GET")
	router.HandleFunc("/certificates", createCertificate).Methods("POST")
	router.HandleFunc("/certificates", getAllCertificates).Methods("GET")
	router.HandleFunc("/certificates/{id}", updateCertificate).Methods("PUT")

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
