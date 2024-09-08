package auth

import "github.com/treenq/treenq/pkg/vel"

var (
	ErrSignIn = vel.Error{
		Code: "SIGN_IN",
	}
	ErrSignInCallback = vel.Error{
		Code: "SIGN_IN_CALLBACK",
	}
	ErrUnauthorized = vel.Error{
		Code: "UNAUTHORIZED",
	}
	ErrGetAccessToken = vel.Error{
		Code: "GET_ACCESS_TOKEN",
	}
)
