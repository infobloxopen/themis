package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const missingStorageMsg = `"Server missing policy storage"`

type storageHandler struct {
	s *Server
}

func (handler *storageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		depth int64
		err   error
	)

	// parse depth
	queryOpt := r.URL.Query()
	if depthOpt, ok := queryOpt["depth"]; ok {
		depthStr := depthOpt[0]
		depth, err = strconv.ParseInt(depthStr, 10, 64)
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
	path := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })[1:]
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
