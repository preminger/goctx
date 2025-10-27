package srv

import (
	"net/http"
)

// Handler is an http handler; boundary should derive ctx := req.Context()
func Handler(w http.ResponseWriter, r *http.Request) {
	_ = w
	_ = r // keep
	inner()
}

// inner should get ctx and comments preserved.
func inner() {}
