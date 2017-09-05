# 05-Target

The example shows policies file with different kinds of target.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p target.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=target.yaml
INFO[0000] Parsing policy                                policy=target.yaml
INFO[0000] Opening service port                          address="0.0.0.0:5555"
INFO[0000] Opening control port                          address="0.0.0.0:5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving decision requests
INFO[0000] Serving control requests
```

In other terminal run pepcli:
```
$ pepcli -i target.requests.yaml
Got 5 requests. Sending...
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "first"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "second"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "third"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "fourth"

- effect: NOTAPPLICABLE
  reason: "Ok"

```

PDP logs:
```
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- x.(String): "test"
- c.(Network): 192.0.2.0/24
- a.(Address): 192.0.2.1
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- r.(string): "first" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- a.(Address): 192.0.2.17
- x.(String): "test"
- c.(Network): 192.0.2.32/28
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- r.(string): "second" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- c.(Network): 192.0.2.32/28
- a.(Address): 192.0.2.33
- x.(String): "test"
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- r.(string): "third" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- x.(String): "test"
- c.(Network): 192.0.2.32/28
- a.(Address): 192.0.3.1
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- r.(string): "fourth" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- a.(Address): 192.0.3.1
- x.(String): "example"
- c.(Network): 192.0.2.32/28
INFO[0003] Returning response
DEBU[0003] Response                                      effect=NOTAPPLICABLE obligation=no attributes reason=Ok
```
