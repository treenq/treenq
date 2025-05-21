package domain

import (
	"context"
	"encoding/json"
	"log" // Assuming standard log is needed for Printf, though slog is available
	"log/slog"
	"net/http"
	"time"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

// IssuePATResponse defines the response body when a PAT is issued.
type IssuePATResponse struct {
	PAT string `json:"pat"`
}

type Handler struct {
	db           Database
	githubClient GithubClient
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	kube         Kube

	kubeConfig string

	oauthProvider   OauthProvider
	jwtIssuer       JwtIssuer
	authRedirectUrl string
	authTtl         time.Duration

	l *slog.Logger
}

func NewHandler(
	db Database,
	githubClient GithubClient,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	kube Kube,
	kubeConfig string,

	oauthProvider OauthProvider,
	jwtIssuer JwtIssuer,
	authRedirectUrl string,
	authTtl time.Duration,

	l *slog.Logger,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		kube:         kube,

		kubeConfig: kubeConfig,

		oauthProvider:   oauthProvider,
		jwtIssuer:       jwtIssuer,
		authRedirectUrl: authRedirectUrl,
		authTtl:         authTtl,
		l:               l,
	}
}

type Database interface {
	// User domain
	////////////////////////
	GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error)

	// Deployment domain
	// ////////////////
	SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error)
	UpdateDeployment(ctx context.Context, def AppDeployment) error
	GetDeployment(ctx context.Context, deploymentID string) (AppDeployment, error)
	GetDeploymentHistory(ctx context.Context, appID string) ([]AppDeployment, error)

	// Github repos domain
	// //////////////////////
	LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error)
	SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubRepos(ctx context.Context, email string) ([]Repository, error)
	GetInstallationID(ctx context.Context, userID string) (string, int, error)
	SaveInstallation(ctx context.Context, userID string, githubID int) (string, error)
	ConnectRepo(ctx context.Context, userID, repoID, branchName string) (Repository, error)
	GetRepoByGithub(ctx context.Context, githubRepoID int) (Repository, error)
	GetRepoByID(ctx context.Context, userID, repoID string) (Repository, error)
	RepoIsConnected(ctx context.Context, repoID string) (bool, error)

	// PAT domain
	CreateUserPAT(ctx context.Context, userID string, encryptedPAT string) error
	GetUserPATByUserID(ctx context.Context, userID string) (string, error) // Assuming it returns encrypted PAT and error
	DeleteUserPAT(ctx context.Context, userID string) error
	GetAllUserPATs(ctx context.Context) ([]string, error) // Assuming it returns a slice of encrypted PATs
}

type GithubClient interface {
	IssueAccessToken(installationID int) (string, error)
	GetUserInstallation(ctx context.Context, displayName string) (int, error)
	ListRepositories(ctx context.Context, installationID int) ([]Repository, error)
	GetBranches(ctx context.Context, installationID int, owner string, repoName string, fresh bool) ([]string, error)
}

type Git interface {
	Clone(url string, installationID int, repoID string, accesstoken string) (GitRepo, error)
}

type Extractor interface {
	ExtractConfig(repoDir string) (tqsdk.Space, error)
}

type DockerArtifactory interface {
	Image(args BuildArtifactRequest) Image
	Build(ctx context.Context, args BuildArtifactRequest, progress *ProgressBuf) (Image, error)
}

type Kube interface {
	DefineApp(ctx context.Context, id string, app tqsdk.Space, image Image) string
	Apply(ctx context.Context, rawConig, data string) error
}

type OauthProvider interface {
	AuthUrl(string) string
	ExchangeUser(ctx context.Context, code string) (UserInfo, error)
}

type JwtIssuer interface {
	GenerateJwtToken(claims map[string]any) (string, error)
}

// IssuePAT handles the creation of a new Personal Access Token for the authenticated user.
func (h *Handler) IssuePAT(w http.ResponseWriter, r *http.Request) {
	// TODO: Get User ID from context (assuming authentication middleware adds it)
	// Example: userID, ok := r.Context().Value("userID").(string)
	// If not ok, return unauthorized error.
	// This part needs to integrate with your existing authentication mechanism.
	userID := "placeholder_user_id" // Replace with actual user ID retrieval

	// Check if user already has a PAT
	_, err := h.db.GetUserPATByUserID(r.Context(), userID)
	if err == nil { // If err is nil, a PAT exists.
		http.Error(w, "User already has an active PAT. Delete the existing one to create a new one.", http.StatusConflict)
		return
	}
	// TODO: More robust error checking needed here, e.g., if it's a different error than "not found"
	// For example, check if errors.Is(err, domain.ErrPATNotFound) or similar specific error

	// Generate new PAT
	plainTextPAT, err := generatePAT() // This function is from pat.go
	if err != nil {
		h.l.Error("Error generating PAT", "error", err)
		http.Error(w, "Failed to generate PAT", http.StatusInternalServerError)
		return
	}

	// Encrypt PAT
	encryptedPAT, err := encryptPAT(plainTextPAT) // This function is from pat.go
	if err != nil {
		h.l.Error("Error encrypting PAT", "error", err)
		http.Error(w, "Failed to encrypt PAT", http.StatusInternalServerError)
		return
	}

	// Store encrypted PAT in database
	err = h.db.CreateUserPAT(r.Context(), userID, encryptedPAT)
	if err != nil {
		h.l.Error("Error storing PAT", "error", err)
		http.Error(w, "Failed to store PAT", http.StatusInternalServerError)
		return
	}

	// Return the plainTextPAT to the user (only this one time)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(IssuePATResponse{PAT: plainTextPAT}); err != nil {
		h.l.Error("Error encoding response", "error", err)
	}
}

// DeletePAT handles the deletion of the authenticated user's PAT.
func (h *Handler) DeletePAT(w http.ResponseWriter, r *http.Request) {
	// TODO: Get User ID from context
	userID := "placeholder_user_id" // Replace with actual user ID retrieval

	err := h.db.DeleteUserPAT(r.Context(), userID)
	if err != nil {
		// TODO: Handle specific errors like "not found" if necessary
		// if errors.Is(err, domain.ErrPATNotFound) {
		// http.Error(w, "No active PAT found for this user", http.StatusNotFound)
		// return
		// }
		h.l.Error("Error deleting PAT", "error", err)
		http.Error(w, "Failed to delete PAT", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ValidatePAT handles the validation of a PAT provided as a query parameter.
// This endpoint is unauthenticated.
func (h *Handler) ValidatePAT(w http.ResponseWriter, r *http.Request) {
	providedKey := r.URL.Query().Get("key")
	if providedKey == "" {
		http.Error(w, "Missing 'key' query parameter", http.StatusBadRequest)
		return
	}

	allEncryptedPATs, err := h.db.GetAllUserPATs(r.Context())
	if err != nil {
		h.l.Error("Error retrieving PATs for validation", "error", err)
		http.Error(w, "Failed to validate PAT", http.StatusInternalServerError)
		return
	}

	isValid := false
	for _, encryptedPAT := range allEncryptedPATs {
		decryptedKey, err := decryptPAT(encryptedPAT) // This function is from pat.go
		if err != nil {
			// Log error but continue, as one key might be corrupted or use an old format
			h.l.Warn("Error decrypting a PAT during validation", "error", err)
			continue
		}
		if decryptedKey == providedKey {
			isValid = true
			break
		}
	}

	if !isValid {
		http.Error(w, "Invalid or expired PAT", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK) // Key is valid
}
