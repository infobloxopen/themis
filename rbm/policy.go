package main

const (
	policyBody = iota
	policyRule
	attrValue
	attrTypeIdx
)

var policies = map[string]map[string][]string{
	"regex": {
		"prefix": {
			regexPrefixPolicy,
			regexPrefixRule,
			prefixValue,
			strType,
		},
		"infix": {
			regexInfixPolicy,
			regexInfixRule,
			infixValue,
			strType,
		},
		"postfix": {
			regexPostfixPolicy,
			regexPostfixRule,
			postfixValue,
			strType,
		},
	},
	"wildcard": {
		"prefix": {
			wcPrefixPolicy,
			wcPrefixRule,
			prefixValue,
			strType,
		},
		"infix": {
			wcInfixPolicy,
			wcInfixRule,
			infixValue,
			strType,
		},
		"postfix": {
			wcPostfixPolicy,
			wcPostfixRule,
			postfixValue,
			strType,
		},
	},
	"mapper": {
		"postfix": {
			mapPostfixPolicy,
			mapPostfixRule,
			mapPostfixValue,
			mapType,
		},
	},
}

const (
	regexPrefixPolicy = `# Regex prefix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      regex-match:
      - val:
          type: string
          content: "^prefix-match-"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	regexPrefixRule = `
  - condition:
      regex-match:
      - val:
          type: string
          content: "^prefix-%d-not-match-"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	prefixValue = "prefix-match-test"

	strType = "string"

	regexInfixPolicy = `# Regex infix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      regex-match:
      - val:
          type: string
          content: "-infix-match-"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	regexInfixRule = `
  - condition:
      regex-match:
      - val:
          type: string
          content: "-infix-%d-not-match-"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	infixValue = "test-infix-match-test"

	regexPostfixPolicy = `# Regex postfix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      regex-match:
      - val:
          type: string
          content: "-postfix-match$"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	regexPostfixRule = `
  - condition:
      regex-match:
      - val:
          type: string
          content: "-postfix-%d-not-match$"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	postfixValue = "test-postfix-match"

	wcPrefixPolicy = `# Wildcard prefix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "prefix-match-*"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	wcPrefixRule = `
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "prefix-%d-not-match-*"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	wcInfixPolicy = `# Wildcard infix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "*-infix-match-*"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	wcInfixRule = `
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "*-infix-%d-not-match-*"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	wcPostfixPolicy = `# Wildcard postfix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:%s
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "*-postfix-match"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`

	wcPostfixRule = `
  - condition:
      wildcard-match:
      - val:
          type: string
          content: "*-postfix-%d-not-match"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	mapPostfixPolicy = `# Mapper postfix policy
attributes:
  x: domain
  c: string

policies:
  alg:
    id: mapper
    map:
      selector:
        uri: "local:content/map"
        path:
        - attr: x
        type: string
  rules:%s
  - id: postfix-match
    effect: Permit
    obligations:
    - c: "rule-match"
`

	mapPostfixRule = `
  - id: "postfix-%d-not-match"
    effect: Permit
    obligations:
    - c: "rule-not-match"`

	mapPostfixValue = "test.example.com"

	mapPostfixContent = `{
	"id": "content",
	"items": {
		"map": {
			"keys": ["domain"],
			"type": "string",
			"data": {
				"example.com": "postfix-match"
			}
		}
	}
}
`

	mapType = "domain"
)
