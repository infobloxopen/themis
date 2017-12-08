package main

const regexPrefixPolicy = `# Regex prefix policy
attributes:
  x: string
  c: string

policies:
  alg: FirstApplicableEffect
  rules:
  - condition:
      regex-match:
      - val:
          type: string
          content: "prefix-match-.*"
      - attr: x
    effect: Permit
    obligations:
    - c: "rule-match"
`
