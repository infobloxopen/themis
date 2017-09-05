# 06-Values

The example shows policies file with different immediate values.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p values.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=values.yaml
INFO[0000] Parsing policy                                policy=values.yaml
INFO[0000] Opening service port                          address="0.0.0.0:5555"
INFO[0000] Opening control port                          address="0.0.0.0:5554"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving decision requests
INFO[0000] Serving control requests
```

In other terminal run pepcli:
```
$ pepcli -i values.requests.yaml
Got 8 requests. Sending...
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "example"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "a"
      type: "address"
      value: "192.0.2.1"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "c"
      type: "network"
      value: "192.0.2.0/28"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "d"
      type: "domain"
      value: "example.net"

    - id: "sd"
      type: "set of domains"
      value: "\"example.com\",\"test.com\""

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "ss"
      type: "set of strings"
      value: "\"first\",\"second\""

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "sn"
      type: "set of networks"
      value: "\"192.0.2.0/28\",\"192.0.2.16/28\""

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "first-rule"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "s"
      type: "string"
      value: "second-rule"

```

PDP logs:
```
...
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "test"
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- s.(string): "example" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "example"
- c.(Network): 192.0.2.0/31
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- a.(address): "192.0.2.1" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "example"
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.13
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- c.(network): "192.0.2.0/28" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- d.(Domain): domain(example.com)
- s.(String): "example"
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.16
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- d.(domain): "example.net"
- sd.(set of domains): "\"example.com\",\"test.com\"" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "first"
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.16
- d.(Domain): domain(example.net)
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- ss.(set of strings): "\"first\",\"second\"" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "third"
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.16
- d.(Domain): domain(example.net)
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- sn.(set of networks): "\"192.0.2.0/28\",\"192.0.2.16/28\"" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.33
- d.(Domain): domain(example.net)
- s.(String): "first-rule"
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- s.(string): "first-rule" reason=Ok
INFO[0003] Validating context
DEBU[0003] Request context                               context=attributes:
- s.(String): "second-rule"
- c.(Network): 192.0.2.2/31
- a.(Address): 192.0.2.33
- d.(Domain): domain(example.net)
INFO[0003] Returning response
DEBU[0003] Response                                      effect=PERMIT obligation=attributes:
- s.(string): "second-rule" reason=Ok
...
```
