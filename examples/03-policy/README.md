# 03-Policy

The example shows policies file with full featured policy.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p policy.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=policy.yaml
INFO[0000] Parsing policy                                policy=policy.yaml
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving control requests
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Serving decision requests
```

In other terminal run pepcli:
```
$ pepcli -i policy.requests.yaml test
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
DEBU[0097] Request context                               context="attributes:
- x.(String): \"test\""
DEBU[0097] Response                                      effect=PERMIT obligation="attributes:
- a.(address): \"192.0.2.1\"" reason=Ok
DEBU[0097] Request context                               context="attributes:
- x.(String): \"example\""
DEBU[0097] Response                                      effect=NOTAPPLICABLE obligation="no attributes" reason=Ok
...
```
