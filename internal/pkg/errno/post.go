package errno

import (
	"net/http"

	"github.com/moweilong/milady/pkg/errorsx"
)

// ErrPostNotFound indicates that the specified blog post was not found.
var ErrPostNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.PostNotFound", Message: "Post not found."}
