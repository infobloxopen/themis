package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type storageHandler struct {
	s *Server
}

func (handler *storageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// sanity check
	root := handler.s.p
	if root == nil {
		io.WriteString(w, `{"error": "Server missing policy storage"}`)
		return
	}

	// parse depth
	queryOpt := r.URL.Query()
	depthOpt, ok := queryOpt["depth"]
	if !ok {
		io.WriteString(w, `{"error": "depth option not specified"}`)
		return
	}
	depthStr := depthOpt[0]
	depth, err := strconv.ParseInt(depthStr, 10, 64)
	if err != nil {
		fmt.Fprintf(w, `{"error": "%v+"}`, err)
		return
	}

	// parse path
	path := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })[1:]
	target, err := root.GetAtPath(path)
	if err != nil {
		fmt.Fprintf(w, `{"error": "%v+"}`, err)
		return
	}

	// dump
	if err = target.MarshalWithDepth(w, int(depth)); err != nil {
		fmt.Fprintf(w, `{"error": "%v+"}`, err)
		return
	}
}
