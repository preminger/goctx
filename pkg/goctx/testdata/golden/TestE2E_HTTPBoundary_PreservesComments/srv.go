package srv

import (
	"context"
	"net/http"
)

// Handler is an http handler; boundary should derive ctx := req.Context()
func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = w
	_ = r // keep
	inner(ctx)
}

// inner should get ctx and comments preserved.
func inner(ctx context.Context) {}
