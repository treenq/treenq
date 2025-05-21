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

	// Rollback deployment
	vel.RegisterHandlerFunc(router, vel.HandlerMeta{
		Path:        "/deployments/{deploymentID}/rollback", // Path with parameter
		Input:       domain.RollbackDeploymentRequest{},   // Input struct for vel to populate
		Output:      domain.RollbackDeploymentResponse{},  // Expected output struct
		Method:      "POST",
		OperationID: "rollbackDeployment", // Unique operation ID
		Middlewares: []vel.Middleware{auth}, // Apply auth middleware
	}, handlers.RollbackDeployment)

	return router
}
