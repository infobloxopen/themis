package yast

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalSelector(v interface{}) (pdp.Expression, boundError) {
	m, err := ctx.validateMap(v, "selector attributes")
	if err != nil {
		return pdp.LocalSelector{}, err
	}

	s, err := ctx.extractString(m, yastTagURI, "selector URI")
	if err != nil {
		return pdp.LocalSelector{}, err
	}

	fmt.Printf("uri string is '%s'\n", s)
	ID, ierr := url.Parse(s)
	if ierr != nil {
		return pdp.LocalSelector{}, newSelectorURIError(s, ierr)
	}

	fmt.Printf("ID is '%v'\n", ID)
	scheme := strings.ToLower(ID.Scheme)
	loc := strings.Split(ID.Opaque, "/")
	fmt.Printf("Scheme is %v, Loc is %v\n", scheme, loc)

	if scheme == "local" {
		if len(loc) != 2 {
			return pdp.LocalSelector{}, newSelectorLocationError(ID.Opaque, s)
		}

		return makeLocalSelector(ctx, loc, m)
	} else if scheme == "pip" {
		if len(loc) != 2 {
			return pdp.PIPSelector{}, newSelectorLocationError(ID.Opaque, s)
		}

		return makePIPSelector(ctx, loc, m)
	}

	return pdp.LocalSelector{}, newUnsupportedSelectorSchemeError(ID.Scheme, s)
}
