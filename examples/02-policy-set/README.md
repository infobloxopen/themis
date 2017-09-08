# 02-Policy-Set

The example shows policies file with full featured policy set.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p policy-set.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=policy-set.yaml
INFO[0000] Parsing policy                                policy=policy-set.yaml
INFO[0000] Opening service port                          address="0.0.0.0:5555"
INFO[0000] Opening control port                          address="0.0.0.0:5554"
INFO[0000] Creating service protocol handler
INFO[0000] Serving decision requests
INFO[0000] Creating control protocol handler
INFO[0000] Serving control requests
```

In other terminal run pepcli:
```
$ pepcli -i policy-set.requests.yaml
Got 4 requests. Sending...
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "a"
      type: "address"
      value: "192.0.2.1"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "a"
      type: "address"
      value: "192.0.2.1"

- effect: NOTAPPLICABLE
  reason: "Ok"

- effect: INDETERMINATEP
  reason: "#07: Failed to process request: #02 (policy set \"Test Policy Set\">hidden policy set>target>any>all>match>equal>second argument>attr(z.String)): Missing attribute"
```

PDP logs:
```
...
INFO[0243] Validating context
DEBU[0243] Request context                               context=attributes:
- z.(String): "example"
- x.(String): "test"
INFO[0243] Returning response
DEBU[0243] Response                                      effect=PERMIT obligation=attributes:
- a.(address): "192.0.2.1" reason=Ok
INFO[0243] Validating context
DEBU[0243] Request context                               context=attributes:
- x.(String): "test"
- z.(String): "test"
INFO[0243] Returning response
DEBU[0243] Response                                      effect=DENY obligation=attributes:
- a.(address): "192.0.2.1" reason=Ok
INFO[0243] Validating context
DEBU[0243] Request context                               context=attributes:
- z.(String): "test"
- x.(String): "example"
INFO[0243] Returning response
DEBU[0243] Response                                      effect=NOTAPPLICABLE obligation=no attributes reason=Ok
INFO[0243] Validating context
DEBU[0243] Request context                               context=attributes:
- x.(String): "test"
INFO[0243] Returning response
DEBU[0243] Response                                      effect=INDETERMINATEP obligation=no attributes reason="#07: Failed to process request: #02 (policy set "Test Policy Set">hidden policy set>target>any>all>match>equal>second argument>attr(z.String)): Missing attribute"
...
```
