# 08-Mapper

The example shows policies file with mapper combining algorithm and local content file.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -p mapper.yaml -j content.json
INFO[0000] Starting PDP server                          
INFO[0000] Loading policy                                policy=mapper.yaml
INFO[0000] Parsing policy                                policy=mapper.yaml
INFO[0000] Opening content                               content=content.json
INFO[0000] Parsing content                               content=content.json
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Creating service protocol handler            
INFO[0000] Creating control protocol handler            
INFO[0000] Serving control requests                     
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Serving decision requests                    
```

In other terminal run pepcli:
```
$ pepcli -i mapper.requests.yaml test
- effect: DENY
  reason: "Ok"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "err"
      type: "string"
      value: "Can't calculate policy id"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "First PermitNet"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "First DenyCom"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "Second PermitCom"

- effect: DENY
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "Second DenyNet"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "External Second"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "p"
      type: "string"
      value: "Internal First"
```

PDP logs:
```
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"Unknown\""
DEBU[0033] Response                                      effect=DENY obligation="no attributes" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- d.(Domain): domain(example.com)"
DEBU[0033] Response                                      effect=DENY obligation="attributes:
- err.(string): \"Can't calculate policy id\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"First\"
- d.(Domain): domain(example.net)"
DEBU[0033] Response                                      effect=PERMIT obligation="attributes:
- p.(string): \"First PermitNet\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"First\"
- d.(Domain): domain(example.com)"
DEBU[0033] Response                                      effect=DENY obligation="attributes:
- p.(string): \"First DenyCom\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"Second\"
- d.(Domain): domain(example.com)"
DEBU[0033] Response                                      effect=PERMIT obligation="attributes:
- p.(string): \"Second PermitCom\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"Second\"
- d.(Domain): domain(example.net)"
DEBU[0033] Response                                      effect=DENY obligation="attributes:
- p.(string): \"Second DenyNet\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"External\""
DEBU[0033] Response                                      effect=PERMIT obligation="attributes:
- p.(string): \"External Second\"" reason=Ok
DEBU[0033] Request context                               context="content:
- content: no tag

attributes:
- p.(String): \"Internal\""
DEBU[0033] Response                                      effect=PERMIT obligation="attributes:
- p.(string): \"Internal First\"" reason=Ok
```
