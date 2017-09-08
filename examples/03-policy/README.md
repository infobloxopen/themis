# 03-Policy

The example shows policies file with full featured policy.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p policy.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=policy.yaml
INFO[0000] Parsing policy                                policy=policy.yaml
INFO[0000] Opening service port                          address="0.0.0.0:5555"
INFO[0000] Opening control port                          address="0.0.0.0:5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving decision requests
INFO[0000] Serving control requests
```

In other terminal run pepcli:
```
$ pepcli -i policy.requests.yaml
Got 2 requests. Sending...
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "a"
      type: "address"
      value: "192.0.2.1"

- effect: NOTAPPLICABLE
  reason: "Ok"
```

PDP logs:
```
...
INFO[0097] Validating context
DEBU[0097] Request context                               context=attributes:
- x.(String): "test"
INFO[0097] Returning response
DEBU[0097] Response                                      effect=PERMIT obligation=attributes:
- a.(address): "192.0.2.1" reason=Ok
INFO[0097] Validating context
DEBU[0097] Request context                               context=attributes:
- x.(String): "example"
INFO[0097] Returning response
DEBU[0097] Response                                      effect=NOTAPPLICABLE obligation=no attributes reason=Ok
...
```
