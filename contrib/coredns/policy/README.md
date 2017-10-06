# policy

*policy* is the Policy Enforcement Point for Themis project implemented as CoreDNS plugin.

## Syntax

~~~ txt
policy {
    endpoint ADDR
    edns0 CODE NAME [SRCTYPE DSTTYPE] [SIZE START END]
    ...
    debug_query_suffix SUFFIX
}
~~~

Valid SRCTYPE are hex (default), bytes, ip.

Valid DSTTYPE depends on Themis PDP implementation, ATM is supported string (default), address.

Params SIZE, START, END is supported only for SRCTYPE = hex.

Set param SIZE to value > 0 enables edns0 option data size check.

Param START and END (last data byte index + 1) allow to get separate part of edns0 option data.

Option debug_query_suffix SUFFIX (should have dot at the end) enables debug query feature.


## Example

~~~ txt
policy {
    endpoint 10.0.0.7:1234
    edns0 0xffee client_id hex string 32 0 32
    edns0 0xffee group_id hex string 32 16 32
    edns0 0xffef uid // equal edns0 0xffef uid hex string
    edns0 0xffea source_ip ip address
    edns0 0xffeb client_name bytes string
    debug_query_suffix debug.
}
~~~

In this case edns0 options with code 0xffee is splitted into two values - client_id (first 16 bytes) and group_id (last 16 bytes), option should have size 32 bytes otherwise client_id and group_id is not parsed.

Dig command example for debug query:
~~~ txt
dig @127.0.0.1 msn.com.debug txt ch
~~~
