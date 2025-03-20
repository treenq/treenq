package router

import (
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/src/domain"
)

func NewRouter(handlers *domain.Handler, auth, githubAuth vel.Middleware, middlewares ...vel.Middleware) *vel.Router {
	router := vel.NewRouter()
	for i := range middlewares {
		router.Use(middlewares[i])
	}

	vel.RegisterHandlerFunc(router, "GET /auth", handlers.GithubAuthHandler)
	vel.RegisterGet(router, "authCallback", handlers.GithubCallbackHandler)

	vel.RegisterPost(router, "githubWebhook", handlers.GithubWebhook, githubAuth)

	// regular authentication handlers
	vel.RegisterPost(router, "info", handlers.Info, auth)
	vel.RegisterPost(router, "getProfile", handlers.GetProfile, auth)
	vel.RegisterPost(router, "getRepos", handlers.GetRepos, auth)
	vel.RegisterPost(router, "connectRepoBranch", handlers.ConnectBranch, auth)

	return router
}
