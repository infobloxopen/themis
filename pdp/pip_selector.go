package pdp

import "fmt"

type PIPSelector struct {
	service   string
	queryType string
	path      []Expression
	t         int
}

func MakePIPSelector(service, queryType string, path []Expression, t int) PIPSelector {
	return PIPSelector{
		service:   service,
		queryType: queryType,
		path:      path,
		t:         t}
}

func (s PIPSelector) GetResultType() int {
	return s.t
}

func (s PIPSelector) describe() string {
	return fmt.Sprintf("PIPselector(%s.%s)", s.service, s.queryType)
}

func (s PIPSelector) getAttributeValue(ctx *Context) (AttributeValue, error) {
	domainStr, err := s.path[0].calculate(ctx)
	if err != nil {
		return undefinedValue, err
	}
	fmt.Printf("domainStr is '%v'\n", domainStr)
	return MakeStringValue("Naughty"), nil
}

func (s PIPSelector) calculate(ctx *Context) (AttributeValue, error) {
	fmt.Printf("s is %v\n", s)

	/*
			item, err := ctx.getContentItem(s.content, s.item)
			fmt.Printf("item is %v\n", item)

			if err != nil {
				return undefinedValue, bindError(err, s.describe())
			}

			if s.t != item.t {
				return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, item.t), s.describe())
			}

		r, err := s.getCategory(s.path, ctx)
		fmt.Printf("r is %v\n", r)

		if err != nil {
			return undefinedValue, bindError(err, s.describe())
		}

		t := r.GetResultType()
		if t != s.t {
			return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, t), s.describe())
		}

		return r, nil
	*/
	return s.getAttributeValue(ctx)
}
