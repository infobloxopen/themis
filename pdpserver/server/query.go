package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/infobloxopen/themis/pdp"
	"github.com/julienschmidt/httprouter"
)

const (
	storageQueryCommands = `PDP Query API

GET /root
GET /storage/<path...>?depth=<depth> [depth is any value of range 0 .. n. defaults to 0]
GET /find/<id>/<start path...> [ensure to end with "/" if start path is root]
`
	maximumRetries  = 100
	maxUint32       = 2147483647
	defaultDepthMsg = "proceeding with depth=0"
)

func queryIndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, storageQueryCommands)
}

func (s *Server) listenQuery() error {
	if len(s.opts.queryEP) <= 0 {
		return nil
	}
	storageQueryRouter := httprouter.New()
	storageQueryRouter.GET("/", queryIndexHandler)
	storageQueryRouter.GET("/root",
		func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
			if s.p == nil {
				fmt.Fprint(w, "policy storage doesn't exist yet")
			} else if id, ok := s.p.Root().GetID(); ok {
				fmt.Fprintf(w, "policy storage has root id %s", strconv.Quote(id))
			} else {
				fmt.Fprint(w, "policy storage root is hidden")
			}
		})
	storageQueryRouter.GET("/storage/*path",
		func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
			if s.p == nil {
				fmt.Fprint(w, "policy storage doesn't exist yet")
				return
			}

			depth := uint64(0)
			path := strings.FieldsFunc(ps.ByName("path"),
				func(r rune) bool { return r == '/' })
			queryOpt := req.URL.Query()
			if depthStr, ok := queryOpt["depth"]; ok && len(depthStr) > 0 {
				parsedDepth, err := strconv.ParseUint(depthStr[0], 10, 32)
				if err == nil {
					depth = parsedDepth
				}
			}

			iter, err := s.p.GetPath(path)
			if err != nil {
				fmt.Fprint(w, err)
			} else {
				treeJson := pdp.GetSubtree(iter, uint(depth))
				// beautify treeJson
				var buf bytes.Buffer
				err := json.Indent(&buf, []byte(treeJson), "", "    ")
				if err != nil {
					fmt.Fprintf(w, "failed to beautify with error %s", err.Error())
				} else {
					fmt.Fprint(w, string(buf.Bytes()))
				}
			}
		})
	storageQueryRouter.GET("/find/:id/*path",
		func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
			if s.p == nil {
				fmt.Fprint(w, "policy storage doesn't exist yet")
				return
			}

			path := strings.FieldsFunc(ps.ByName("path"),
				func(r rune) bool { return r == '/' })
			id := ps.ByName("id")
			iter, err := s.p.GetPath(path)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}

			tPath, _, err := pdp.PathQuery(iter, id)
			if err != nil {
				fmt.Fprint(w, err)
			} else {
				path = append(path, tPath...)
				for i, p := range path {
					path[i] = strconv.Quote(p)
				}
				fmt.Fprintf(w, "found %s @<%s>", id, strings.Join(path, "/"))
			}
		})
	var err error
	go func() {
		err = http.ListenAndServe(s.opts.queryEP, storageQueryRouter)
	}()
	for retries := 0; retries < maximumRetries; retries-- {
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			return fmt.Errorf("Serving storage query failed: %s", err)
		}
		response, err := http.Get(s.opts.queryEP)
		if err == nil {
			if response.StatusCode != 200 {
				response.Body.Close()
				return fmt.Errorf("Debug http API returns %d", response.StatusCode)
			}
			response.Body.Close()
			return nil
		}
		if response != nil {
			response.Body.Close()
		}
	}
	return fmt.Errorf("Cannot reach endpoint %s", s.opts.queryEP)
}
