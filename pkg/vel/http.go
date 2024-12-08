package vel

import (
	"context"
	"net/http"
)

func Redirect(ctx context.Context, url string, status int) {
	r := RequestFromContext(ctx)
	w := WriterFromContext(ctx)

	http.Redirect(w, r, url, status)
}
