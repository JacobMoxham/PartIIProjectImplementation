package http

import (
	"bytes"
	"net/http"
)

func privacyAwareHandler(policy policies.DataPolicy) http.HandlerFunc {
	return policy.defaultHandler
}

func main() {
	finalHandler := http.HandlerFunc(final)

	http.Handle("/", enforceXMLHandler(finalHandler))
	http.ListenAndServe(":3000", nil)
}

func final(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
