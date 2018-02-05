# 06-Values

The example shows policies file with numerical functions.

Run pdpserver using policy file:
```
$ pdpserver -v 3 -pfmt yaml -p numerical.yaml 
INFO[0000] Starting PDP server                          
INFO[0000] Loading policy                                policy=numerical.yaml
INFO[0000] Parsing policy                                policy=numerical.yaml
INFO[0000] Opening service port                          address=":5555"
INFO[0000] Opening control port                          address=":5554"
INFO[0000] Creating service protocol handler            
INFO[0000] Creating control protocol handler            
INFO[0000] Serving control requests                     
```

In other terminal run pepcli:
```
$ pepcli -i numerical.requests.yaml test
- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "float"
      value: "2"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "float"
      value: "10"

- effect: PERMIT
  reason: "Ok"
  obligation:
    - id: "r"
      type: "float"
      value: "3.2"
```

PDP logs:
```
...
INFO[0000] Serving decision requests                    
DEBU[0014] Request context                               context=attributes:
- targetVal.(Float): 5
- actualVal.(Float): 5
DEBU[0014] Response                                      effect=PERMIT obligation=attributes:
- r.(float): "2" reason=Ok
DEBU[0014] Request context                               context=attributes:
- actualVal.(Float): 100
- targetVal.(Float): 5
DEBU[0014] Response                                      effect=PERMIT obligation=attributes:
- r.(float): "10" reason=Ok
DEBU[0014] Request context                               context=attributes:
- actualVal.(Float): 16
- targetVal.(Float): 5
DEBU[0014] Response                                      effect=PERMIT obligation=attributes:
- r.(float): "3.2" reason=Ok
...
```
