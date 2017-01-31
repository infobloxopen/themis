package pdp

import (
	"encoding/json"
	"io/ioutil"
)

func (ctx *yastCtx) loadJSONFile(path string) (interface{}, error) {
	f, err := findAndOpenFile(path, ctx.dataDir)
	if err != nil {
		return nil, ctx.errorf("opening file %s: %v", path, err)
	}

	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, ctx.errorf("reading file %s: %v", path, err)
	}

	m := make(map[string]interface{})
	mErr := json.Unmarshal(d, &m)
	if mErr == nil {
		return m, nil
	}

	a := make([]interface{}, 0)
	aErr := json.Unmarshal(d, &a)
	if aErr == nil {
		return a, nil
	}

	return nil, ctx.errorf("unmarshaling file %s to object: %v, to array: %v", path, mErr, aErr)
}

func (ctx *yastCtx) unmarshalInclude(k interface{}, v interface{}) error {
	ID, err := ctx.validateString(k, "include id")
	if err != nil {
		return err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	_, ok := ctx.includes[ID]
	if ok {
		return nil
	}

	path, err := ctx.validateString(v, "path to file")
	if err != nil {
		return err
	}

	content, err := ctx.loadJSONFile(path)
	if err != nil {
		return err
	}

	ctx.includes[ID] = content
	return nil
}

func (ctx *yastCtx) unmarshalIncludes(m map[interface{}]interface{}, ext map[string]interface{}) error {
	if ext == nil {
		ctx.includes = make(map[string]interface{})
	} else {
		ctx.includes = ext
	}

	incls, err := ctx.extractMap(m, yastTagInclude, "include")
	if err != nil {
		return err
	}

	ctx.pushNodeSpec(yastTagInclude)
	defer ctx.popNodeSpec()

	for k, v := range incls {
		err := ctx.unmarshalInclude(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
