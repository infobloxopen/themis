# 07-Selector

The example shows policies file with selector and local content file.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p selector.yaml -j content.json
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=selector.yaml
INFO[0000] Parsing policy                                policy=selector.yaml
INFO[0000] Opening content                               content=content.json
INFO[0000] Parsing content                               content=content.json
INFO[0000] Opening service port                          address="0.0.0.0:5555"
INFO[0000] Opening control port                          address="0.0.0.0:5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving decision requests
INFO[0000] Serving control requests
```

In other terminal run pepcli:
```
$ pepcli -i selector.requests.yaml
Got 4 requests. Sending...
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "Good"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "Bad"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "Good"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "Bad"

```

PDP logs:
```
...
INFO[0280] Validating context
DEBU[0280] Request context                               context=content:
- content: no tag

attributes:
- d.(Domain): domain(example.com)
- a.(Address): 192.0.2.18
INFO[0280] Returning response
DEBU[0280] Response                                      effect=PERMIT obligation=attributes:
- s.(string): "Good" reason=Ok
INFO[0280] Validating context
DEBU[0280] Request context                               context=content:
- content: no tag

attributes:
- a.(Address): 2001:db8:1000::1
- d.(Domain): domain(example.com)
INFO[0280] Returning response
DEBU[0280] Response                                      effect=DENY obligation=attributes:
- s.(string): "Bad" reason=Ok
INFO[0280] Validating context
DEBU[0280] Request context                               context=content:
- content: no tag

attributes:
- d.(Domain): domain(test.com)
- a.(Address): 192.0.2.50
INFO[0280] Returning response
DEBU[0280] Response                                      effect=PERMIT obligation=attributes:
- s.(string): "Good" reason=Ok
INFO[0280] Validating context
DEBU[0280] Request context                               context=content:
- content: no tag

attributes:
- d.(Domain): domain(test.com)
- a.(Address): 2001:db8:3000::1
INFO[0280] Returning response
DEBU[0280] Response                                      effect=DENY obligation=attributes:
- s.(string): "Bad" reason=Ok
...
```
