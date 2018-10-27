package middleware

import (
	"net/http"
)

func privacyAwareHandler(policy DataPolicy) http.HandlerFunc {
	return policy.defaultHandler
}
