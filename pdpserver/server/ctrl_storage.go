package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const (
	queryCmd          = "query"
	missingStorageMsg = `"Server missing policy storage"`
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

type storageHandler struct {
	s *Server
}

func (handler *storageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		depth uint64
		err   error
	)
	path := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })
	if len(path) == 0 || path[0] != queryCmd {
		http.Error(w, usage, 404)
		return
	}

	// parse depth
	queryOpt := r.URL.Query()
	if depthOpt, ok := queryOpt["depth"]; ok {
		depthStr := depthOpt[0]
		depth, err = strconv.ParseUint(depthStr, 10, 31)
		if err != nil {
			http.Error(w, strconv.Quote(err.Error()), 400)
			return
		}
	}

	// sanity check
	root := handler.s.p
	if root == nil {
		http.Error(w, missingStorageMsg, 404)
		return
	}

	// parse path
	path = path[1:] // remove queryCmd
	target, err := root.GetAtPath(path)
	if err != nil {
		var errCode int
		if _, ok := err.(*pdp.PathNotFoundError); ok {
			errCode = 404
		} else {
			errCode = 500
		}
		http.Error(w, strconv.Quote(err.Error()), errCode)
		return
	}

	// dump
	if err = target.MarshalWithDepth(w, int(depth)); err != nil {
		http.Error(w, strconv.Quote(err.Error()), 500)
		return
	}
}
