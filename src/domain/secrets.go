package domain

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	// Assuming other necessary internal packages like "store", "services/cdk" will be used by Handler methods
)

// SetSecretRequest defines the structure for the request body of HandleSetSecret.
type SetSecretRequest struct {
	SecretKey   string `json:"secret_key"`
	SecretValue string `json:"secret_value"`
}

// ViewSecretResponse defines the structure for the response body of HandleViewSecret.
type ViewSecretResponse struct {
	SecretValue string `json:"secret_value"`
}

// HandleSetSecret handles the creation or update of a repository secret.
// It stores the secret key in the database and the encrypted secret value in Kubernetes.
func (h *Handler) HandleSetSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoID, ok := vars["repoID"]
	if !ok {
		http.Error(w, "Repository ID is missing in URL path", http.StatusBadRequest)
		return
	}

	var req SetSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.SecretKey == "" {
		http.Error(w, "secret_key cannot be empty", http.StatusBadRequest)
		return
	}
	if req.SecretValue == "" { // Assuming secret value also cannot be empty
		http.Error(w, "secret_value cannot be empty", http.StatusBadRequest)
		return
	}

	// Define Kubernetes namespace and secret object name
	// Based on the prompt: namespace = repoID, secretObjectName = "treenq-secrets"
	namespace := repoID
	secretObjectName := "treenq-secrets"

	// 1. Store the secret key metadata in the database
	// Note: CreateRepositorySecretKey only stores the key, not the value.
	// It's idempotent for (repository_id, secret_key) due to UNIQUE constraint.
	if err := h.store.CreateRepositorySecretKey(r.Context(), repoID, req.SecretKey); err != nil {
		// We might want to check if the error is due to a unique constraint violation (key already exists)
		// For now, any error from the store is treated as a server error.
		// If the key already exists, this call might be a no-op or return an error depending on implementation.
		// Assuming it's okay if it already exists, or the method handles idempotency.
		// If it returns a "duplicate key" error and we want to treat "set" as "create or update",
		// we might ignore that specific error here. For now, let's assume any error is critical.
		http.Error(w, "Failed to save secret key metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Store the actual secret value in Kubernetes
	err := h.kube.StoreSecretValue(r.Context(), h.kubeConfig, namespace, secretObjectName, req.SecretKey, req.SecretValue)
	if err != nil {
		http.Error(w, "Failed to store secret in Kubernetes: "+err.Error(), http.StatusInternalServerError)
		// Potentially: if storing in K8s fails, should we roll back the DB write?
		// For simplicity now, we don't. A more robust solution might use a two-phase commit or a cleanup mechanism.
		return
	}

	w.WriteHeader(http.StatusCreated) // Or http.StatusOK if we consider it an update as well
	json.NewEncoder(w).Encode(map[string]string{"status": "secret set successfully"})
}

// HandleGetSecrets handles listing all secret keys for a repository.
func (h *Handler) HandleGetSecrets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoID, ok := vars["repoID"]
	if !ok {
		http.Error(w, "Repository ID is missing in URL path", http.StatusBadRequest)
		return
	}

	keys, err := h.store.GetRepositorySecretKeys(r.Context(), repoID)
	if err != nil {
		// This could be sql.ErrNoRows if no keys are found, or another DB error.
		// GetRepositorySecretKeys is expected to return an empty slice and no error if no keys are found.
		// So, any error here is likely a genuine server-side issue.
		http.Error(w, "Failed to retrieve secret keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If keys is nil (e.g., store returns nil slice for no keys), make it an empty slice for JSON encoding
	if keys == nil {
		keys = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		// Log this error, as the headers are already sent.
		// Consider a centralized error logging/handling mechanism.
		// For now, we can't send another http.Error.
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleViewSecret handles retrieving the actual value of a specific secret key.
func (h *Handler) HandleViewSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoID, okRepo := vars["repoID"]
	secretKey, okSecretKey := vars["secretKey"]

	if !okRepo || !okSecretKey {
		http.Error(w, "Repository ID or Secret Key is missing in URL path", http.StatusBadRequest)
		return
	}

	// 1. Check if the secret key is registered in the database for this repository
	exists, err := h.store.RepositorySecretKeyExists(r.Context(), repoID, secretKey)
	if err != nil {
		http.Error(w, "Failed to check secret key existence: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Secret key not found for this repository", http.StatusNotFound)
		return
	}

	// 2. Retrieve the secret value from Kubernetes
	// Based on the prompt: namespace = repoID, secretObjectName = "treenq-secrets"
	namespace := repoID
	secretObjectName := "treenq-secrets"

	value, err := h.kube.GetSecretValue(r.Context(), h.kubeConfig, namespace, secretObjectName, secretKey)
	if err != nil {
		// This could be due to the K8s secret object or the specific key not being found,
		// or a problem decoding it. If the DB said it exists, this indicates a potential
		// inconsistency or an issue during the StoreSecretValue phase.
		// We might differentiate between "not found in K8s" (which is an inconsistency)
		// and other K8s errors. For now, treat most K8s errors here as server errors,
		// but specific "not found" errors could also be 404.
		// The GetSecretValue implementation already returns a distinct error for not found.
		// Let's assume if the error message contains "not found", it's a 404.
		// This is a bit heuristic; ideally, GetSecretValue would return typed errors.
		// For now, let's use a generic message for 500 and be specific for 404 if GetSecretValue implies it.
		// Based on GetSecretValue's current error messages:
		// "secret ... not found..." or "secret key ... not found..."
		if err.Error().Contains("not found") {
			http.Error(w, "Secret not found in Kubernetes: "+err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve secret from Kubernetes: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ViewSecretResponse{SecretValue: value}); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
