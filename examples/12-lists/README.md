# 12-Lists

The example shows policies file with lists of strings functions.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -pfmt yaml -p lists.yaml
INFO[0000] Starting PDP server
INFO[0000] Loading policy                                policy=lists.yaml
INFO[0000] Parsing policy                                policy=lists.yaml
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Opening storage port                          address=":5552"
INFO[0000] Creating service protocol handler
INFO[0000] Creating control protocol handler
INFO[0000] Serving control requests
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Serving decision requests
```

In other terminal run pepcli:
```
$ pepcli -i lists.requests.yaml test
- effect: Deny
  obligation:
    - id: "i"
      type: "list of strings"
      value: ""

- effect: Deny
  obligation:
    - id: "i"
      type: "list of strings"
      value: ""

- effect: Permit
  obligation:
    - id: "i"
      type: "list of strings"
      value: "\"foo\""

- effect: Permit
  obligation:
    - id: "i"
      type: "list of strings"
      value: "\"foo\",\"bar\""
      
```

PDP logs:
```
...
DEBU[0008] Request context                               context="attributes:
- ls.(List of Strings): []"
DEBU[0008] Response                                      effect=Deny obligations="attributes:
- i.(list of strings): \"\"" reason="<nil>"
DEBU[0008] Request context                               context="attributes:
- ls.(List of Strings): [\"\", \"baz\"]"
DEBU[0008] Response                                      effect=Deny obligations="attributes:
- i.(list of strings): \"\"" reason="<nil>"
DEBU[0008] Request context                               context="attributes:
- ls.(List of Strings): [\"\", \"\", ...]"
DEBU[0008] Response                                      effect=Permit obligations="attributes:
- i.(list of strings): \"\\\"foo\\\"\"" reason="<nil>"
DEBU[0008] Request context                               context="attributes:
- ls.(List of Strings): [\"\", \"\", ...]"
DEBU[0008] Response                                      effect=Permit obligations="attributes:
- i.(list of strings): \"\\\"foo\\\",\\\"bar\\\"\"" reason="<nil>"
...
```
