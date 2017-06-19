package main

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

const originalPolicies = `
attributes:
  str_attr1: string
  addr_attr1: address
  str_attr2: string
policies:
  id: root
  alg: FirstApplicableEffect
  policies:
    - id: P1
      alg:
        id: Mapper
        map:
          selector:
            type: String
            path:
              - attr: str_attr1
            content: content1
      policies:
        - id: P1P1
          alg: FirstApplicableEffect
          rules:
            - id: P1P1R1
              effect: Permit
            - id: P1P1R2
              effect: Deny
        - id: P1P2
          alg: FirstApplicableEffect
          rules:
            - id: P1P2R1
              effect: Permit
            - id: P1P2R2
              effect: Deny
            - id: P1P2R3
              effect: Permit
    - id: P2
      alg:
        id: Mapper
        map:
          selector:
            type: String
            path:
              - ID1
              - ID11
              - ID111
              - attr: addr_attr1
            content: content2
      policies:
        - id: P2P1
          alg: FirstApplicableEffect
          rules:
            - id: P2P1R1
              effect: Permit
        - id: P2P2
          alg: FirstApplicableEffect
          rules:
            - id: P2P2R1
              effect: Permit
            - id: P2P2R2
              effect: Deny
    - id: P3
      alg:
        alg: FirstApplicableEffect
        id: Mapper
        map:
          selector:
            type: set of strings
            content: content3
            path:
              - attr: str_attr2
      policies:
        - id: P3P1
          alg: FirstApplicableEffect
          rules:
            - id: P3P1R1
              effect: Permit
`

func originalContent() map[string]interface{} {
	return map[string]interface{}{
		"content1": map[string]interface{}{
			"c1_key1": "c1_val1",
			"c1_key2": "c1_val2",
		},
		"content2": map[string]interface{}{
			"ID1": map[string]interface{}{
				"ID11": map[string]interface{}{
					"ID111": map[string]interface{}{
						"9.9.9.9/32": "c2_val1",
						"1.1.1.1/16": "c2_val2",
					},
				},
				"ID12": map[string]interface{}{
					"ID121": map[string]interface{}{
						"10.10.10.10/32": "c2_val1",
					},
				},
			},
		},
		"content3": map[string]interface{}{
			"ADDRS": []interface{}{"example1.com", "example2.com"},
		},
	}
}

func testServer(ctx pdp.YastCtx, p pdp.EvaluableType, inc map[string]interface{}) *Server {
	return &Server{
		Policy:           p,
		Includes:         inc,
		AffectedPolicies: map[string]pdp.ContentPolicyIndexItem{},
		Ctx:              ctx,
	}
}

func testFindPoliciesItem(path []string, i int, e pdp.EvaluableType) (interface{}, bool) {
	id := path[i]
	switch e := e.(type) {
	case *pdp.PolicySetType:
		for _, v := range e.Policies {
			if v.GetID() == id {
				if len(path) > i+1 {
					return testFindPoliciesItem(path, i+1, v)
				}
				return v, true
			}
		}
	case *pdp.PolicyType:
		for _, r := range e.Rules {
			if r.ID == id {
				return r, true
			}
		}
	}

	return nil, false
}

func assertPoliciesItemExists(t *testing.T, path []string, e pdp.EvaluableType) {
	if _, ok := testFindPoliciesItem(path, 0, e); !ok {
		t.Fatalf("Can't find '%v' policies item", path)
	}
}

func assertPoliciesItemDoesNotExist(t *testing.T, path []string, e pdp.EvaluableType) {
	if _, ok := testFindPoliciesItem(path, 0, e); ok {
		t.Fatalf("Found '%v' unexpected policies item", path)
	}
}

func checkSelectorContent(t *testing.T, s *pdp.SelectorType, val interface{}) {
	c := s.Content
	switch s.DataType {
	case pdp.DataTypeString:
		cmap := c.(map[string]interface{})
		vals := val.([]string)

		if len(cmap) != len(vals) {
			t.Fatalf("Expected '%d' num of items but got '%d'", len(vals), len(cmap))
		}
		for _, k := range vals {
			if _, ok := cmap[k]; !ok {
				t.Fatalf("Expected '%s' entry in selector content '%+v'", k, cmap)
			}
		}
	case pdp.DataTypeSetOfStrings:
		cmap := c.(map[string]interface{})
		vmap := val.(map[string][]string)
		for k, vals := range vmap {
			vt := cmap[k].(pdp.AttributeValueType)
			cmapmap := vt.Value.(map[string]int)
			if len(cmapmap) != len(vals) {
				t.Fatalf("Expected '%d' num of items but got '%d'", len(vals), len(cmapmap))
			}
			for _, k := range vals {
				if _, ok := cmapmap[k]; !ok {
					t.Fatalf("Expected '%s' entry in selector content '%+v'", k, c)
				}
			}
		}
	default:
		t.Fatalf("Unexpected selector content type '%T' type'", c)
	}
}

func assertCheckPolicySelector(t *testing.T, e interface{}, val interface{}) {
	switch e := e.(type) {
	case *pdp.PolicySetType:
		mpca := e.AlgParams.(*pdp.MapperPCAParams)
		s := mpca.Argument.(*pdp.SelectorType)
		checkSelectorContent(t, s, val)
	case *pdp.PolicyType:
		mrca := e.AlgParams.(*pdp.MapperRCAParams)
		s := mrca.Argument.(*pdp.SelectorType)
		checkSelectorContent(t, s, val)
	default:
		t.Fatalf("Unexpected evaluable type '%T'", e)
	}
}

func testFindContentItem(path []string, i int, c interface{}) (interface{}, bool) {
	id := path[i]
	switch c := c.(type) {
	case map[string]interface{}:
		for k, v := range c {
			if k == id {
				if len(path) > i+1 {
					return testFindContentItem(path, i+1, v)
				}
				return v, true
			}
		}
	case []interface{}:
		for _, v := range c {
			if v == id {
				return v, true
			}
		}
	case string:
		if id == c {
			return c, true
		}
	}

	return nil, false
}

func assertContentItemExists(t *testing.T, path []string, c interface{}) {
	if _, ok := testFindContentItem(path, 0, c); !ok {
		t.Fatalf("Can't find '%v' content item", path)
	}
}

func assertContentItemDoesNotExist(t *testing.T, path []string, c interface{}) {
	if _, ok := testFindContentItem(path, 0, c); ok {
		t.Fatalf("Found '%v' unexpected content item", path)
	}
}

const policiesPatch1 = `
- op: add
  path:
    - root
    - P1
    - P1P2
  entity:
    id: P1P2R4
    effect: Permit

- op: delete
  path:
    - root
    - P1
    - P1P1
    - P1P1R1

- op: delete
  path:
    - root
    - P3
- op: add
  path:
    - root
  entity:
    id: P4
    alg:
      id: Mapper
      map:
        selector:
          type: String
          path:
            - ID1
            - ID11
            - ID111
            - attr: addr_attr1
          content: content2
    policies:
      - id: P4P1
        alg: FirstApplicableEffect
        rules:
          - id: P4P1R1
            effect: Permit
`

func TestPoliciesPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	content := originalContent()
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := testServer(ctx, p, content)

	pp, err := s.copyAndPatchPolicies([]byte(policiesPatch1), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check policiesPatch1:add[1] operation.
	for _, v := range []string{"P1P2R1", "P1P2R2", "P1P2R3"} {
		assertPoliciesItemExists(t, []string{"P1", "P1P2", v}, p)
	}
	assertPoliciesItemDoesNotExist(t, []string{"P1", "P1P2", "P1P2R4"}, p)

	for _, v := range []string{"P1P2R1", "P1P2R2", "P1P2R3", "P1P2R4"} {
		assertPoliciesItemExists(t, []string{"P1", "P1P2", v}, pp)
	}

	// Check policiesPatch1:delete[2] operation.
	for _, v := range []string{"P1P1R1", "P1P1R2"} {
		assertPoliciesItemExists(t, []string{"P1", "P1P1", v}, p)
	}

	assertPoliciesItemDoesNotExist(t, []string{"P1", "P1P1", "P1P1R1"}, pp)
	assertPoliciesItemExists(t, []string{"P1", "P1P1", "P1P1R2"}, pp)

	// Check policiesPatch1:delete[3] operation.
	assertPoliciesItemExists(t, []string{"P3"}, p)
	assertPoliciesItemDoesNotExist(t, []string{"P3"}, pp)

	// Check policiesPatch1:add[4] operation.
	assertPoliciesItemDoesNotExist(t, []string{"P4"}, p)
	assertPoliciesItemExists(t, []string{"P4"}, pp)
}

const content1Patch1 = `
[{"op": "add",
  "path": ["c1_key3"],
  "entity": "c1_val3"},

 {"op": "delete",
  "path": ["c1_key2"]}
]
`

const content2Patch1 = `
[{"op": "add",
  "path": ["ID1", "ID11", "ID111", "9.9.9.9/32"],
  "entity": "c2_newval1"},

 {"op": "add",
  "path": ["ID1", "ID11", "ID111", "2.2.2.2/32"],
  "entity": "c2_val3"},

 {"op": "delete",
  "path": ["ID1", "ID11", "ID111", "1.1.1.1/16"]}
]
`
const content2Patch2 = `
[{"op": "delete",
  "path": ["ID1", "ID12"]},

 {"op": "add",
  "path": ["ID1", "ID13"],
  "entity": {"ID131": {"100.100.100.100/32": "c2_val1"}}}
]
`

const content2Patch3 = `
[{"op": "delete",
  "path": ["ID1", "NOTEXIST"]}
]
`

func TestContentPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	content := originalContent()
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := testServer(ctx, p, content)

	_, err = s.patchContent([]byte(content1Patch1), "content1")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check content1Patch1 operations.
	assertContentItemExists(t, []string{"content1", "c1_key1"}, content)
	assertContentItemExists(t, []string{"content1", "c1_key3"}, content)
	assertContentItemDoesNotExist(t, []string{"content1", "c1_key2"}, content)

	_, err = s.patchContent([]byte(content2Patch1), "content2")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check content2Patch1:add[1] operation.
	assertContentItemExists(t, []string{"content2", "ID1", "ID11", "ID111", "9.9.9.9/32", "c2_newval1"}, content)

	// Check content2Patch1:add[2] operation.
	assertContentItemExists(t, []string{"content2", "ID1", "ID11", "ID111", "2.2.2.2/32", "c2_val3"}, content)
	assertContentItemExists(t, []string{"content2", "ID1", "ID11", "ID111", "9.9.9.9/32", "c2_newval1"}, content)

	// Check content2Patch1:delete[3] operation.
	assertContentItemDoesNotExist(t, []string{"content2", "ID1", "ID11", "ID111", "1.1.1.1/16"}, content)

	_, err = s.patchContent([]byte(content2Patch2), "content2")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check content2Patch2:delete[1] operation.
	assertContentItemDoesNotExist(t, []string{"content2", "ID1", "ID12"}, content)

	// Check content2Patch2:add[2] operation
	assertContentItemExists(t, []string{"content2", "ID1", "ID13", "ID131", "100.100.100.100/32", "c2_val1"}, content)

	_, err = s.patchContent([]byte(content2Patch3), "content2")
	if err == nil {
		t.Fatal("Expected error but got OK")
	}
}

const content1Patch2 = `
[{"op": "add",
  "path": ["c1_key3"],
  "entity": "c1_val3"}
]
`

const content3Patch2 = `
[{"op": "add",
  "path": ["ADDRS"],
  "entity": ["example1.com", "example2.com", "test1.com"]},
 {"op": "delete",
  "path": ["ADDRS", "example1.com"]}
]
`

func TestPoliciesUpdateOnContentPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	content := originalContent()
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := testServer(ctx, p, content)

	_, err = s.patchContent([]byte(content1Patch2), "content1")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	assertContentItemExists(t, []string{"content1", "c1_key3"}, content)

	s.Ctx.Reset()
	s.Ctx.SetContent(content)

	pp, err := s.copyAndPatchPolicies([]byte(""), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check policy update on content1Patch2:add[1] operation.
	op1, _ := testFindPoliciesItem([]string{"P1"}, 0, p)
	assertCheckPolicySelector(t, op1, []string{"c1_key1", "c1_key2"})

	pp1, _ := testFindPoliciesItem([]string{"P1"}, 0, pp)
	assertCheckPolicySelector(t, pp1, []string{"c1_key1", "c1_key2", "c1_key3"})

	s.Ctx.Reset()
	s.Ctx.SetContent(content)

	_, err = s.patchContent([]byte(content3Patch2), "content3")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check content3Patch2:add[1] operation.
	for _, v := range []string{"example2.com", "test1.com"} {
		assertContentItemExists(t, []string{"content3", "ADDRS", v}, content)
	}
	assertContentItemDoesNotExist(t, []string{"content3", "ADDRS", "example1.com"}, content)

	pp, err = s.copyAndPatchPolicies([]byte(""), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check policy update on content3Patch2:add[1] operation.
	pp3, _ := testFindPoliciesItem([]string{"P3"}, 0, pp)
	assertCheckPolicySelector(t, pp3, map[string][]string{
		"ADDRS": {"example2.com", "test1.com"}})
}

const policiesPatch2 = `
- op: delete
  path:
    - root
    - P2
`

const content2Patch4 = `
[{"op": "add",
  "path": ["ID1", "ID11", "ID111", "2.2.2.2/32"],
  "entity": "c2_val3"}
]
`

func TestPoliciesAndContentPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	content := originalContent()
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := testServer(ctx, p, content)

	_, err = s.patchContent([]byte(content2Patch4), "content2")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	_, err = s.copyAndPatchPolicies([]byte(policiesPatch2), content)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}
}
