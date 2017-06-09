package pdp

import (
	"fmt"
	"strings"
)

type YASTError struct {
	breadcrumbs []string
	message     string
}

func (e *YASTError) Error() string {
	return fmt.Sprintf("%s> %s", strings.Join(e.breadcrumbs, ">"), e.message)
}

func (ctx YastCtx) errorf(format string, a ...interface{}) error {
	breadcrumbs := make([]string, 0)
	if ctx.nodeSpec != nil && len(ctx.nodeSpec) > 0 {
		breadcrumbs = ctx.nodeSpec
	}

	return &YASTError{breadcrumbs, fmt.Sprintf(format, a...)}
}

func (ctx YastCtx) makeRootError(m map[interface{}]interface{}) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, fmt.Sprintf("%v", k))
	}

	return ctx.errorf("Expected attribute definitions, includes and policies but got: %s", strings.Join(keys, ", "))
}
