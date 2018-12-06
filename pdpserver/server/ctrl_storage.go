package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const (
	queryCmd          = "query"
	readonlyMsg       = "This endpoint is read only. Only GET method is allowed"
	missingStorageMsg = "Server missing policy storage"
	usage             = `PDP storage traversal API:
Description: This API displays the ord, id, target, obligation, and algorithm
information of specific nodes/subtree in the pdp storage tree. The subtree root
is identified by the path parameter, and the depth of the subtree is specified
by the depth parameter.

Parameters:

	path    Is an optional url parameter in the form of 'id1/id2/.../id3'
			where id2 is the child of id1.
			Select the subtree of the right-most id in this path
			(id3 in above case).
            By default, the path is empty (the root node is selected).

    depth   Is an optional query string parameter taking a positive integer.
			Display a subtree with at most depth specified.
			E.g.: depth=1 displays the selected root and its children.
            By default, the depth is 0 (only display the selected node).

GET /query/<path>?depth=<depth>`
)

func handleQuery(w http.ResponseWriter, store *pdp.PolicyStorage,
	path []string, urlQuery url.Values) {
	// sanity check
	if store == nil {
		http.Error(w, missingStorageMsg, http.StatusNotFound)
		return
	}

	var (
		depth uint64
		err   error
	)

	// parse depth
	if depthOpt, ok := urlQuery["depth"]; ok {
		depth, err = strconv.ParseUint(depthOpt[0], 10, 31)
		if err != nil {
			http.Error(w, strconv.Quote(err.Error()), http.StatusBadRequest)
			return
		}
	}

	// parse path
	target, err := store.GetAtPath(path)
	if err != nil {
		var errCode int
		if _, ok := err.(*pdp.PathNotFoundError); ok {
			errCode = http.StatusNotFound
		} else {
			errCode = http.StatusInternalServerError
		}
		http.Error(w, strconv.Quote(err.Error()), errCode)
		return
	}

	// dump
	if err = target.MarshalWithDepth(w, int(depth)); err != nil {
		http.Error(w, strconv.Quote(err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type storageHandler struct {
	s *Server
}

func (handler *storageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, readonlyMsg, http.StatusMethodNotAllowed)
		return
	}

	path := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })
	if len(path) == 0 {
		http.Error(w, usage, http.StatusNotFound)
		return
	}

	urlQuery := r.URL.Query()
	cmd := path[0]
	resourcePath := path[1:]

	switch cmd {
	case queryCmd:
		handleQuery(w, handler.s.p, resourcePath, urlQuery)
	default:
		http.Error(w, fmt.Sprintf("Unknown resource %s\n%s", cmd, usage), http.StatusNotFound)
	}
}
