package resources

import (
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/src/domain"
)

func NewRouter(handlers *domain.Handler, auth, githubAuth vel.Middleware, middlewares ...vel.Middleware) *vel.Router {
	router := vel.NewRouter()
	for i := range middlewares {
		router.Use(middlewares[i])
	}

	// auth is an endpoint contain redirect, therefore it must be GET
	vel.RegisterHandlerFunc(router, vel.HandlerMeta{
		Input:       struct{}{},
		Output:      domain.GithubCallbackResponse{},
		Method:      "GET",
		OperationID: "auth",
	}, handlers.GithubAuthHandler)
	vel.RegisterGet(router, "authCallback", handlers.GithubCallbackHandler)

	// vcs webhooks
	vel.RegisterPost(router, "githubWebhook", handlers.GithubWebhook, githubAuth)

	// treenq api
	vel.RegisterPost(router, "logout", handlers.Logout, auth)
	vel.RegisterPost(router, "info", handlers.Info, auth)
	vel.RegisterPost(router, "getProfile", handlers.GetProfile, auth)
	vel.RegisterPost(router, "getRepos", handlers.GetRepos, auth)
	vel.RegisterPost(router, "getBranches", handlers.GetBranches, auth)
	vel.RegisterPost(router, "syncGithubApp", handlers.SyncGithubApp, auth)
	vel.RegisterPost(router, "connectRepoBranch", handlers.ConnectBranch, auth)
	vel.RegisterPost(router, "deploy", handlers.Deploy, auth)
	vel.RegisterPost(router, "getDeployment", handlers.GetDeployment, auth)
	vel.RegisterGet(router, "getBuildProgress", handlers.GetBuildProgress, auth)

	// PAT routes
	// GET /hook?key=... (ValidatePAT - public)
	// Assuming ValidatePAT doesn't require specific input/output structs for vel.RegisterGet
	// If it does, vel.RegisterHandlerFunc would be more appropriate.
	// For now, let's assume it fits the simple GET pattern or vel.RegisterGet is flexible.
	vel.RegisterGet(router, "hook", handlers.ValidatePAT) // No 'auth' middleware

	// POST /hook (IssuePAT - authenticated)
	vel.RegisterPost(router, "hook", handlers.IssuePAT, auth)

	// DELETE /hook (DeletePAT - authenticated)
	// Using RegisterHandlerFunc as there's no specific RegisterDelete
	vel.RegisterHandlerFunc(router, vel.HandlerMeta{
		Input:       struct{}{}, // Define input struct if DeletePAT expects a body
		Output:      struct{}{}, // Define output struct if DeletePAT returns a body
		Method:      "DELETE",
		OperationID: "deletePat", // Or an appropriate operation ID
		Path:        "hook",      // The path for this route
	}, handlers.DeletePAT, auth)

	return router
}
