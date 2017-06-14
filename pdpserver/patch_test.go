package main

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

const originalPolicies = `
attributes:
  str_attr1: string
  addr_attr1: address
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
              - ID12
              - ID123
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
`

var originalContent = map[string]interface{}{
	"content1": map[string]interface{}{
		"c1_key1": "c1_val1",
		"c1_key2": "c1_val2",
	},
	"content2": map[string]interface{}{
		"ID1": map[string]interface{}{
			"ID12": map[string]interface{}{
				"ID123": map[string]interface{}{
					"9.9.9.9/32": "c2_val1",
					"1.1.1.1/16": "c2_val2",
				},
			},
		},
	},
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
`

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
  "path": ["ID1", "ID12", "ID123", "9.9.9.9/32"],
  "entity": "c2_newval1"},

 {"op": "add",
  "path": ["ID1", "ID12", "ID123", "2.2.2.2/32"],
  "entity": "c1_val3"},

 {"op": "delete",
  "path": ["ID1", "ID12", "ID123", "1.1.1.1/16"]}
]
`

func TestPiliciesPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), originalContent)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := &Server{Policy: p, Includes: originalContent, ctx: ctx}

	pp, err := s.copyAndPatchPolicies([]byte(policiesPatch1), originalContent)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	// Check policiesPatch1:append[1] operation.
	oP1P2Rules := p.(*pdp.PolicySetType).Policies[0].(*pdp.PolicySetType).Policies[1].(*pdp.PolicyType).Rules
	pP1P2Rules := pp.(*pdp.PolicySetType).Policies[0].(*pdp.PolicySetType).Policies[1].(*pdp.PolicyType).Rules

	if len(oP1P2Rules) != 3 {
		t.Fatalf("Expected %d rules for original 'root/P1/P1P2' policy but got %d", 3, len(oP1P2Rules))
	}

	if len(pP1P2Rules) != 4 {
		t.Fatalf("Expected %d rules for patched 'root/P1/P1P2' policy but got %d", 4, len(pP1P2Rules))
	}

	if r := pP1P2Rules[3]; r.ID != "P1P2R4" {
		t.Fatalf("Expected %s new rule for patched 'root/P1/P1P2' policy but got %s", "P1P2R4", r.ID)
	}

	// Check policiesPatch1:delete[2] operation.
	oP1P1Rules := p.(*pdp.PolicySetType).Policies[0].(*pdp.PolicySetType).Policies[0].(*pdp.PolicyType).Rules
	pP1P1Rules := pp.(*pdp.PolicySetType).Policies[0].(*pdp.PolicySetType).Policies[0].(*pdp.PolicyType).Rules

	if len(oP1P1Rules) != 2 {
		t.Fatalf("Expected %d rules for original 'root/P1/P1P1' policy but got %d", 2, len(oP1P1Rules))
	}

	if len(pP1P1Rules) != 1 {
		t.Fatalf("Expected %d rules for patched 'root/P1/P1P1' policy but got %d", 1, len(pP1P1Rules))
	}

	if r := pP1P1Rules[0]; r.ID != "P1P1R2" {
		t.Fatalf("Expected %s new rule for patched 'root/P1/P1P1' policy but got %s", "P1P2R2", r.ID)
	}
}

func TestContentPatches(t *testing.T) {
	ctx := pdp.NewYASTCtx("")
	p, err := ctx.UnmarshalYAST([]byte(originalPolicies), originalContent)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	s := &Server{Policy: p, Includes: originalContent, ctx: ctx}

	c1, err := s.patchContent([]byte(content1Patch1), "content1")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	content1 := c1.(map[string]interface{})
	if len(content1) != 2 {
		t.Fatalf("Expected %d items in patched 'content1' but got %d", 2, len(content1))
	}

	c2, err := s.patchContent([]byte(content2Patch1), "content2")
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	content2 := c2.(map[string]interface{})
	id123 := content2["ID1"].(map[string]interface{})["ID12"].(map[string]interface{})["ID123"].(map[string]interface{})
	if len(id123) != 2 {
		t.Fatalf("Expected %d items in patched 'content1/ID1/ID12/ID123' but got %d", 2, len(id123))
	}
}
