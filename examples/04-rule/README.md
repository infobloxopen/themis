# 04-Rule

The example shows policies file with full featured rule.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p rule.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=rule.yaml
INFO[0000] Parsing policy                                policy=rule.yaml
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving control requests
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Serving decision requests
```

In other terminal run pepcli:
```
$ pepcli -i rule.requests.yaml test
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "a"
      type: "address"
      value: "192.0.2.1"

- effect: NOTAPPLICABLE
  reason: "Ok"

- effect: NOTAPPLICABLE
  reason: "Ok"

- effect: NOTAPPLICABLE
  reason: "Ok"

```

PDP logs:
```
...
DEBU[0034] Request context                               context="attributes:
- b.(Boolean): false
- x.(String): \"test\"
- c.(Network): 192.0.2.16/28"
DEBU[0034] Response                                      effect=PERMIT obligation="attributes:
- a.(address): \"192.0.2.1\"" reason=Ok
DEBU[0034] Request context                               context="attributes:
- x.(String): \"test\"
- c.(Network): 192.0.2.16/28
- b.(Boolean): true"
DEBU[0034] Response                                      effect=NOTAPPLICABLE obligation="no attributes" reason=Ok
DEBU[0034] Request context                               context="attributes:
- x.(String): \"test\"
- c.(Network): 192.0.2.0/24
- b.(Boolean): false"
DEBU[0034] Response                                      effect=NOTAPPLICABLE obligation="no attributes" reason=Ok
DEBU[0034] Request context                               context="attributes:
- x.(String): \"example\""
DEBU[0034] Response                                      effect=NOTAPPLICABLE obligation="no attributes" reason=Ok
...
```
