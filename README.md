# Themis
Themis represents a set of tools for managing and enforcing security policies along with framework to create such tools:
- **pdp** - Policy Decision Point (core component of Themis);
- **pdpserver** - standalone application server which runs PDP;
- **proto**, **pdp-service**, **pdp-control** - gRPC protocol definitions and implementations;
- **pep** - golang client package for "service" protocol (Policy Enforcement Point or PEP);
- **pepcli** - CLI application which implements simple PEP;
- **pdpctr-client** - golang client package for "control" protocol (Policy Administration Point or PAP);
- **papcli** - CLI application which implements simple PAP;
- **egen** - error processing code generator (development tool).

Themis design is inspired by eXtensible Access Control Markup Language (XACML) **[XACML-V3.0]**.

# Policy Decision Point
Policy Decision Point or PDP (according to **[XACML-V3.0]**) is an entity that evaluates applicable policy and renders an authorization decision. Themis provides PDP as a golang package.

## Policy Evaluation
To make a decision PDP evaluates policies it has on request **context**. A request **context** represents set of **attributes** together with local **content** (additional data which can be used by policies during the evaluation). Resulting decision consists of **effect**, **status** (**reason**) and set of **obligations**. Decision effect can be:
- **Deny** - request is denied;
- **Permit** - request is permitted;
- **Not Applicable** - no policy applicable to the request;
- **Indeterminate** - PDP can't evaluate particular effect;
- **IndeterminateD** - PDP can't evaluate effect but if it could it would be **Deny**;
- **IndeterminateP** - PDP can't evaluate effect but if it could it would be **Permit**;
- **IndeterminateDP** - PDP can't evaluate effect but if it could it would be only **Deny** or **Permit**;
In case of any **Indeterminate** effect **status** contains textual representation of an issue.

Some application may require more details for particular decision. For example if application can write log it may need a flag attached to decision which says when to do it. These details can be delivered as **obligations**. The **obligations** are set of attributes like in request context. Each attribute has name, type and value. Attribute name is arbitrary string (request context requires pair of attribute name and type to be unique). Following types are defined:
- **boolean**;
- **string**;
- **address** - IPv4 or IPv6 address;
- **network** - IPv4 or IPv6 network address;
- **domain** - domain name;
- **set of strings** - ordered set of strings;
- **set of domains** - set of domains (unordered);
- **set of networks** - set of IPv4 or IPv6 network addresses (unordered);
- **list of strings**. 

**Boolean** value is accepted as "1", "t", "T", "TRUE", "true", "True", "0", "f", "F", "FALSE", "false", "False" and serialized to "true" and "false". **Address** accepted in dotted decimal ("192.0.2.1") form or in IPv6 ("2001:db8::68") form and serialized respectively. **Network** is accepted as a CIDR notation IP address and prefix (for example "192.0.2.0/24" or "2001:db8::/32"). **Domain** name is accepted as string of labels separated by dots (string is converted from punycode to ASCII and validated with regualr expression "^[-.\_A-Za-z0-9]+$" (**TODO**: need to rework according to RFC1035, 2181 and 4343). **Set of strings**, **set of domains**, **set of networks** and **list of strings** aren't accepted in request context but can appear in responce's obligations as comma separated list of values.

## Policies
PDP uses YAML based language (YAML Abstract Syntax Tree or YAST) to define **policies** and specifically constructed JSON to define local **content**.

### Root
Any **policies** definition consists of attributes (optional) and policies (required) sections. Attributes section contains set of pairs attribute name and type. For example:
```yaml
# Permit if x is "test" otherwise Not Applicable
attributes:
  x: string

policies:
  alg: FirstApplicableEffect
  target:
  - equal:
    - attr: x
    - val:
        type: string
        content: "test"
  rules:
  - effect: Permit
```

```yaml
# All permit policy (without "attributes" section)
policies:
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
```

Policies section contains root **policy** or **policy set**. **Policy** holds rules under its "rules" field while **policy set** is able to contain both inner policies or policy sets under its "policies" field.

### Policy Set
Policy Set holds set of **policies** or inner **policy sets** and defines how to combine them. It has following fields:
- **id** - policy id (optional, if not defined policy is hidden);
- **target** - target expression which defines if policy set is applicable to request (optional, if not defined policy set is applicable to any request);
- **policies** - set of inner policies and policy sets;
- **alg** - policy combining algorithm (any of **FirstApplicableEffect**, **DenyOverrides** and **Mapper**);
- **obligations** - set of obligations (optional).

Example of policy set with all its fields (it contains one hidden policy set and one hidden policy):
```yaml
# Policy set with all its fields
attributes:
  x: string
  a: address
  z: string

policies:
  id: "Test Policy Set"
  target: # x == "test"
  - equal:
    - attr: x
    - val:
        type: string
        content: "test"
  alg: FirstApplicableEffect
  policies:
  - alg: FirstApplicableEffect
    target:
    - equal: # z == "example"
      - attr: z
      - val:
          type: string
          content: "example"
    policies:
    - alg: FirstApplicableEffect
      rules:
      - effect: Permit
  - alg: FirstApplicableEffect
    rules:
    - effect: Deny
  obligations:
  - a: "192.0.2.1"
```

### Policy
Policy stores set of rules and defines how to combine them. It has following fields:
- **id** - policy id (optional, if not defined policy is hidden);
- **target** - target expression which defines if policy is applicable to request (optional, if not defined policy is applicable to any request);
- **rules** - set of rules;
- **alg** - rule combining algorithm (the same as for policy set);
- **obligations** - set of obligations (optional).

Here is an example of policy with all fields defined (it contains one hidden rule):
```yaml
# Policy with all its fields
attributes:
  x: string
  a: address

policies:
  id: "Test Policy"
  target: # x == "test"
  - equal:
    - attr: x
    - val:
        type: string
        content: "test"
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
  obligations:
  - a: "192.0.2.1"
```

### Rule
Rule defines decision effect. Possible fields of a rule:
- **id** - rule id (optional, if not defined policy is hidden);
- **target** - target expression (optional);
- **condition** - any boolean expression which together with target defines if rule is applicable to the request (optional, if not defined rule is applicable when target matches);
- **effect** - **Deny** or **Permit**;
- **obligations** - set of obligations.

For example a rule with all fields:
```yaml
# Rule with all its fields
attributes:
  x: string
  a: address
  n: network
  b: boolean

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Policy"
    target: # x == "test"
    - equal:
      - attr: x
      - val:
          type: string
          content: "test"
    condition: # not (n contains 192.0.2.1 or b)
      not:
        or:
        - contains:
          - attr: n
          - val:
              type: address
              content: "192.0.2.1"
        - attr: b
    effect: Permit
    obligations:
    - a: "192.0.2.1"
```

### Target
Any particular policy set or policy or rule is applicable only if request matches its target. Target is a list of **any** expressions. **Any** expression is a list of **all** expressions and **all** expression is a list of match expression. Match expression is a boolean expression of two arguments. One of arguments should be a request attribute and other should be a immediate value. Only two functions can represent match expression **equal** and **contains**. If list of match expressions for particular **all** expression contains single element **all** keyword can be dropped. Similarly if list of **all** expressions for particular **any** expression consists of one element **any** keyword can be dropped.

Request matches target when all **any** expressions match (if one or more of **any** expression doesn't match, target also doesn't match). **Any** expression matches request if one or more of its **all** expressions match the request (if all **all** expressions don't match, **any** expression doesn't match as well). And similarly to target **all** expression matches if all its inner expressions match as well. If during target evaluation error occurs the policy set, policy or rule effect becomes **indeterminate** (if rule effect is permit it is **indeterminateP** if deny - **indeterminateD** for policy and policy set kind of **indeterminate** depends on combining algorithm (see below).

Below an example of policy with different kinds of targets:
```yaml
# Target examples
attributes:
  r: string
  x: string
  a: address
  c: network

policies:
  alg: FirstApplicableEffect
  rules:
  - target: # ((x == test and c contains address(192.0.2.1)) or
            #  x == example) and
            # (network(192.0.2.0/28) contains a or network(192.0.2.16/28) contains a)
    - any:
      - all:
        - equal:
          - attr: x
          - val:
              type: string
              content: "test"
        - contains:
          - attr: c
          - val:
              type: address
              content: 192.0.2.1
      - equal:
        - attr: x
        - val:
            type: string
            content: "example"
    - any:
      - contains:
        - val:
            type: network
            content: 192.0.2.0/28
        - attr: a
      - contains:
        - val:
            type: network
            content: 192.0.2.16/28
        - attr: a
    effect: Permit
    obligations:
    - r: first
  - target: # (x == test or x == example) and (network(192.0.2.0/28) contains a or network(192.0.2.16/28) contains a)
    - any:
      - equal:
        - attr: x
        - val:
            type: string
            content: "test"
      - equal:
        - attr: x
        - val:
            type: string
            content: "example"
    - any:
      - contains:
        - val:
            type: network
            content: 192.0.2.0/28
        - attr: a
      - contains:
        - val:
            type: network
            content: 192.0.2.16/28
        - attr: a
    effect: Permit
    obligations:
    - r: second
  - target: # x == test and network(192.0.2.0/24) contains a
    - equal:
      - attr: x
      - val:
          type: string
          content: "test"
    - contains:
      - val:
          type: network
          content: 192.0.2.0/24
      - attr: a
    effect: Permit
    obligations:
    - r: third
  - target: # x == test
    - equal:
      - attr: x
      - val:
          type: string
          content: "test"
    effect: Permit
    obligations:
    - r: fourth
```

### Condition
Condition is rule field which can be any boolean expression (for example see above "Rule with all its fields"). Following functions available to make such expression:
- **equal** - expects two string arguments (result is true if the arguments are equal)
- **contains**:
  - string contains substring - expects two string arguments first is a string to search in and second is a substring to serach for;
  - network constains address;
  - set of strings contains string;
  - set of networks contains address;
  - set of domains contains doman;
- **not** - boolean not (expects boolean as its single argument);
- **and**, **or** - boolean and and or (expect booleans as its arguments (requires at least one).

In any expression attribute can be referred with **attr** keyword and immediate value with **val** keyword (see below). There is special **selector** expression which is described below.

### Immediate Value
Immediate value can be reffered with **val** keyword and has fields:
- **type** - value type (any of available types);
- **content** - value data.
For **obligations** only data itself requires as type can be derived from attribute definition.

Example of policy with all possible values:
```yaml
# All values example
attributes:
  b: boolean
  s: string
  a: address
  c: network
  d: domain
  ss: set of strings
  sn: set of networks
  sd: set of domains
  ls: list of strings

policies:
  alg: DenyOverrides
  policies:
  - alg: FirstApplicableEffect
    rules:
    - target: # string
      - equal:
        - attr: s
        - val:
            type: string
            content: test
      effect: Permit
      obligations:
      - s: example
    - target: # address
      - contains:
        - attr: c
        - val:
            type: address
            content: 192.0.2.1
      effect: Permit
      obligations:
      - a: 192.0.2.1
    - target: # network
      - contains:
        - attr: a
        - val:
            type: network
            content: 192.0.2.0/28
      effect: Permit
      obligations:
      - c: 192.0.2.0/28
    - condition: # domain and set of domains
        and:
        - contains:
          - val:
              type: set of domains
              content:
              - test.com
              - example.com
          - attr: d
        - contains:
          - val:
              type: set of domains
              content:
              - test.com
              - example.com
          - val:
              type: domain
              content: test.com
      effect: Permit
      obligations:
      - d: example.net
      - sd:
        - test.com
        - example.com
    - target: # set of strings
      - contains:
        - attr: s
        - val:
            type: set of strings
            content:
            - first
            - second
      effect: Permit
      obligations:
      - ss:
        - first
        - second
    - target: # set of networks
      - contains:
        - attr: a
        - val:
            type: set of networks
            content:
            - 192.0.2.0/28
            - 192.0.2.16/28
      effect: Permit
      obligations:
      - sn:
        - 192.0.2.0/28
        - 192.0.2.16/28
  - alg: # list of strings
      id: Mapper
      alg: FirstApplicableEffect
      map:
        val:
          type: list of strings
          content:
          - first
          - second
    rules:
    - target:
      - equal:
        - attr: s
        - val:
            type: string
            content: first-rule
      effect: Permit
      obligations:
      - s: first-rule
    - target:
      - equal:
        - attr: s
        - val:
            type: string
            content: second-rule
      effect: Permit
      obligations:
      - s: second-rule
```

### Selector
Selector expression is an expression to access additionally supplied data. Selector uses **uri** field to locate source of such data. Currently only "local" URI schema is supported which defines local selector.

Local selector uses local content data (see below) and has following fields:
- **uri** - URI of local content ("local:&lt;content-id&gt;/&lt;content-item-id&gt;);
- **path** - defines path to data in local content (optional, if not set selector extracts immediate value from content item). Path represents a list of expressions. It should match to content item keys (see below). Selector calculates path expressions one by one and extracts value from next mapping step of content item until reaches desired value;
- **type** - type of data in local content (any of available types).

Example of local selector:
```yaml
# Selector example
attributes:
  d: domain
  a: address
  s: string

policies:
  alg: FirstApplicableEffect
  rules:
  - target:
    - contains:
      - selector:
          uri: "local:content/domain-addresses"
          path:
          - val:
              type: string
              content: good
          - attr: d
          type: set of networks
      - attr: a
    effect: Permit
    obligations:
    - s: Good
  - target:
    - contains:
      - selector:
          uri: "local:content/domain-addresses"
          path:
          - val:
              type: string
              content: bad
          - attr: d
          type: set of networks
      - attr: a
    effect: Deny
    obligations:
    - s: Bad
```

Content for the example:
```json
{
  "id": "content",
  "items": {
    "domain-addresses": {
      "keys": ["string", "domain"],
      "type": "set of networks",
      "data": {
        "good": {
          "example.com": ["192.0.2.16/28", "192.0.2.32/28"],
          "test.com": ["192.0.2.48/28", "192.0.2.64/28"]
        },
        "bad": {
          "example.com": ["2001:db8:1000::/40", "2001:db8:2000::/40"],
          "test.com": ["2001:db8:3000::/40", "2001:db8:4000::/40"]
        }
      }
    }
  }
}
```

### Local Content
Local content is a set of content items (see example above). It's identified by id field which can be any string with no slash character (`/`). Each content item also has id (key of "items" JSON object) and following fields:
- **keys** - list of types of nested maps (optional, if not present data should contain immediate value of type);
- **type** - any possible type;
- **data** - list of nested maps with keys of mentioned types or immediate value of given type.

Local content supports string map (key type "string"), domain map (key type "domain") and network map (key type "network" or "address"). Selector expectes string expression as path item for string map, domain - for domain map and address or network - for network map ("address" expression is allowed even if content key is "network" and vice verse). 

# References
**[XACML-V3.0]** *eXtensible Access Control Markup Language (XACML) Version 3.0.* 22 January 2013. OASIS Standard. http://docs.oasis-open.org/xacml/3.0/xacml-3.0-core-spec-os-en.html.
