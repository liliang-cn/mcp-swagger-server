package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Pet represents a pet in our store
type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag,omitempty"`
	Age  int    `json:"age,omitempty"`
}

// NewPet represents a pet to be created (without ID)
type NewPet struct {
	Name string `json:"name"`
	Tag  string `json:"tag,omitempty"`
	Age  int    `json:"age,omitempty"`
}

// SearchCriteria represents search parameters
type SearchCriteria struct {
	Name   string `json:"name,omitempty"`
	Tag    string `json:"tag,omitempty"`
	MinAge int    `json:"minAge,omitempty"`
	MaxAge int    `json:"maxAge,omitempty"`
}

// In-memory pet store
type PetStore struct {
	pets   map[int64]Pet
	nextID int64
	mutex  sync.RWMutex
}

var store = &PetStore{
	pets:   make(map[int64]Pet),
	nextID: 1,
	mutex:  sync.RWMutex{},
}

func main() {
	// Initialize with some sample data
	initSampleData()

	// Create router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/v2/pets", listPetsHandler).Methods("GET")
	r.HandleFunc("/v2/pets", createPetHandler).Methods("POST")
	r.HandleFunc("/v2/pets/{petId}", getPetHandler).Methods("GET")
	r.HandleFunc("/v2/pets/{petId}", updatePetHandler).Methods("PUT")
	r.HandleFunc("/v2/pets/{petId}", deletePetHandler).Methods("DELETE")
	r.HandleFunc("/v2/pets/search", searchPetsHandler).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// CORS middleware
	r.Use(corsMiddleware)

	fmt.Println("🐾 Local Petstore API Server")
	fmt.Println("================================")
	fmt.Println("Server starting on port 4538...")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /v2/pets           - List all pets")
	fmt.Println("  POST   /v2/pets           - Create a new pet")
	fmt.Println("  GET    /v2/pets/{id}      - Get pet by ID")
	fmt.Println("  PUT    /v2/pets/{id}      - Update pet")
	fmt.Println("  DELETE /v2/pets/{id}      - Delete pet")
	fmt.Println("  POST   /v2/pets/search    - Search pets")
	fmt.Println("  GET    /health            - Health check")
	fmt.Println()
	fmt.Println("Swagger UI available at: http://localhost:4538/swagger/")
	fmt.Println("API Base URL: http://localhost:4538")

	// Start server
	log.Fatal(http.ListenAndServe(":4538", r))
}

func initSampleData() {
	samplePets := []Pet{
		{ID: 1, Name: "Buddy", Tag: "dog", Age: 3},
		{ID: 2, Name: "Mittens", Tag: "cat", Age: 2},
		{ID: 3, Name: "Goldie", Tag: "fish", Age: 1},
		{ID: 4, Name: "Charlie", Tag: "dog", Age: 5},
		{ID: 5, Name: "Luna", Tag: "cat", Age: 3},
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	for _, pet := range samplePets {
		store.pets[pet.ID] = pet
		if pet.ID >= store.nextID {
			store.nextID = pet.ID + 1
		}
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"server":    "local-petstore-api",
		"version":   "1.0.0",
	})
}

func listPetsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limitStr := query.Get("limit")
	tags := query["tags"]

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var pets []Pet
	for _, pet := range store.pets {
		// Filter by tags if specified
		if len(tags) > 0 {
			found := false
			for _, tag := range tags {
				if strings.EqualFold(pet.Tag, tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		pets = append(pets, pet)
	}

	// Apply limit
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			if len(pets) > limit {
				pets = pets[:limit]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pets)
}

func createPetHandler(w http.ResponseWriter, r *http.Request) {
	var newPet NewPet
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if newPet.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	pet := Pet{
		ID:   store.nextID,
		Name: newPet.Name,
		Tag:  newPet.Tag,
		Age:  newPet.Age,
	}

	store.pets[store.nextID] = pet
	store.nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pet)
}

func getPetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petIdStr := vars["petId"]

	petId, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid pet ID", http.StatusBadRequest)
		return
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	pet, exists := store.pets[petId]
	if !exists {
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pet)
}

func updatePetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petIdStr := vars["petId"]

	petId, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid pet ID", http.StatusBadRequest)
		return
	}

	var newPet NewPet
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if newPet.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.pets[petId]; !exists {
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}

	updatedPet := Pet{
		ID:   petId,
		Name: newPet.Name,
		Tag:  newPet.Tag,
		Age:  newPet.Age,
	}

	store.pets[petId] = updatedPet

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPet)
}

func deletePetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	petIdStr := vars["petId"]

	petId, err := strconv.ParseInt(petIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid pet ID", http.StatusBadRequest)
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.pets[petId]; !exists {
		http.Error(w, "Pet not found", http.StatusNotFound)
		return
	}

	delete(store.pets, petId)
	w.WriteHeader(http.StatusNoContent)
}

func searchPetsHandler(w http.ResponseWriter, r *http.Request) {
	var criteria SearchCriteria
	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var results []Pet
	for _, pet := range store.pets {
		// Filter by name
		if criteria.Name != "" && !strings.Contains(strings.ToLower(pet.Name), strings.ToLower(criteria.Name)) {
			continue
		}

		// Filter by tag
		if criteria.Tag != "" && !strings.EqualFold(pet.Tag, criteria.Tag) {
			continue
		}

		// Filter by age range
		if criteria.MinAge > 0 && pet.Age < criteria.MinAge {
			continue
		}
		if criteria.MaxAge > 0 && pet.Age > criteria.MaxAge {
			continue
		}

		results = append(results, pet)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}