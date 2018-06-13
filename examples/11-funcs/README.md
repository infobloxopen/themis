# 11-Funcs

The example shows how to use functions such as **try** and **concat**.

## Policy with function

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p funcs.yaml
INFO[0000] Starting PDP server                          
INFO[0000] Loading policy                                policy=funcs.yaml
INFO[0000] Parsing policy                                policy=funcs.yaml
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Opening storage port                          address=":5552"
INFO[0000] Creating service protocol handler            
INFO[0000] Creating control protocol handler            
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Serving decision requests                    
INFO[0000] Serving control requests                     
```

In other terminal run pepcli:
```
$ pepcli -i funcs.requests.yaml test
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "test"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "string"
      value: "default"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "ls"
      type: "list of strings"
      value: "\"one\",\"two\",\"three\",\"first\",\"second\",\"third\""

```

PDP logs:
```
...
DEBU[0002] Request context                               context="attributes:
- func.(String): \"try\"
- x.(String): \"test\""
DEBU[0002] Response                                      effect=PERMIT obligation="attributes:
- r.(string): \"test\"" reason=Ok
DEBU[0002] Request context                               context="attributes:
- func.(String): \"try\""
DEBU[0002] Response                                      effect=PERMIT obligation="attributes:
- r.(string): \"default\"" reason=Ok
DEBU[0002] Request context                               context="attributes:
- func.(String): \"concat\""
DEBU[0002] Response                                      effect=PERMIT obligation="attributes:
- ls.(list of strings): \"\\\"one\\\",\\\"two\\\",\\\"three\\\",\\\"first\\\",\\\"second\\\",\\\"third\\\"\"" reason=Ok
...
```
