# policy

*policy* is the Policy Enforcement Point (PEP) for Themis project implemented as CoreDNS plugin.

## Syntax

~~~ txt
policy {
    endpoint ADDR_1 ADDR_2 ... ADDR_N
    streams COUNT
    connection_timeout TIMEOUT
    max_request_size [[auto] SIZE]
    max_response_attributes auto | COUNT
    cache [TTL [SIZE]]

    edns0 CODE NAME [SRCTYPE] [SIZE START END]
    validation1 ATTR_1 ATTR_2 ... ATTR_N
    validation2 ATTR_1 ATTR_2 ... ATTR_N
    default_decision ATTR_1 ATTR_2 ... ATTR_N
    metrics ATTR_1 ATTR_2 ... ATTR_N
    dnstap LEVEL ATTR_1 ATTR_2 ... ATTR_N

    passthrough SUFFIX_1 SUFFIX_2 ... SUFFIX_N
    debug_query_suffix SUFFIX
    debug_id ID
    log
}
~~~

Option **endpoint** defines addresses of PDP servers.

Option **streams** sets number of gRPC streams to be used in each PDP connection.

Option **connection_timeout** sets timeout for query validation when no PDP server is available. Negative value or "no" keyword means wait forever. This is default behavior. With zero timeout validation fails instantly if there is no PDP server. The option works only if gRPC streams number is greater than 0.

Option **max_request_size** sets maximum buffer size in bytes for serialized request. Too high limit makes the plugin to allocate unnecessary memory while too small can lead to buffer overflow errors on validation. With setting `auto` plugin automatically allocates required amount of bytes for each request. In case of both `auto` and *SIZE* defined, the *SIZE* doesn't limit request buffer but used for cache allocations. 

Option **max_response_attributes** defines the limit of attribute number expected in PDP response. If value is `auto` the appropriate buffer for all PDP response attributes is allocated automatically.

Option **cache** enables decision cache. The default value for *TTL* is 10 minutes. *SIZE* (in megabytes) limits the memory cache can use. If *SIZE* is not provided the cache can grow until application crashes due to out of memory.

Option **edns0** is used for parsing edns0 options into PDP attributes, option with code *CODE* is parsed as attribute with name *NAME*. *CODE* can be defined as octal, decimal or hexadecimal value. Hexadecimal numbers should start with prefix `0x`, e.g `0xfff5`, octal numbers should start with `0`, e.g. `0177765`. *SRCTYPE* defines encoding if edns0 data, valid values are `hex` (default), `bytes`, `ip`. Params *SIZE*, *START*, *END* is supported only for *SRCTYPE* = `hex`. Setting param *SIZE* to value > 0 enables edns0 option data size check. Param *START* and *END* (last data byte index + 1) allows picking out a particular part of edns0 option data into a separate attribute. Option **edns0** can be used repeatedly to define several ends0 attributes

Option **validation1** defines a set of attributes to be sent to PDP for validation before resolving domain name.

Option **validation2** defines a set of attributes to be sent to PDP for validation after resolving domain name.

Option **default_decision** defines default values for some attributes in case if PDP request was failed.

Option **metrics** defines the attributes for which metric counters to be generated. The metric counters hold the number of recently received queries per attribute/value and look like below.

~~~ txt
coredns_policy_recent_queries{attribute="uid",value="9e868487da91153c"} 153
~~~

The counter is decremented when queries get expired. The default expiration period is 1 minute. If no new query is received during the expiration period for the given attribute/value then the related counter is removed to save memory.

Option **dnstap** defines the attributes to be included in extra field of DNStap message if those attributes are available. The option can be used repeatedly with different values of *LEVEL* parameter to define different sets of dnstap attributes. Value of *LEVEL* should be in range [1..3]. The level of dnstap logging for particular DNS message is defined by `log` attribute which could be got from PDP response or predefined in plugin configuration.

The *ATTR_N* parameters have the syntax like `<name>[=<value>]` where value can be one of quoted string, number or IP address. See examples below.

Option **passthrough** defines set of domain name suffixes. Domain that has one of these suffixes is resolved without validation. Each suffix should have dot at the end.

Option **debug_query_suffix** enables debug query feature. Debug query returns some debug information about query processing in text representation. *SUFFIX* should have dot at the end. Debug query uses protocol CHAOS and TXT resource record. See below an example of dig command for debug query.

~~~ txt
dig @127.0.0.1 test.com.debug txt ch
~~~

Option **debug_id** defines a string to be sent with debug response to identify a CoreDNS instance.

Option **log** enables logging PDP request and response attributes

## Predefined attributes

Policy plugin recognizes the following predefined attributes:

 - *domain_name* (type string) - initialized with domain name taken from DNS query.

 - *dns_qtype* (type integer) - initialized with numeric code of query type taken from DNS query. For example, code 1 corresponds to type A, code 28 corresponds to type AAAA, code 5 corresponds to CNAME.

 - *source_ip* (type address) - initialized with IP address from which the DNS query was received.

 - *address* (type address) - assigned with resolved IP address corresponding to query domain name (the first A or AAAA record in DNS response).

 - *policy_action* (type integer) - can be assigned with a value defined in default_decision option or with a value taken from PDP response. The valid values are 1 - Drop, 2 - Allow, 3 - Block, 4 - Redirect, 5 - Refuse.

 - *redirect_to* (type string) - can be assigned with a value defined in default_decision option or with a value taken from PDP response. The valid formats are string representation of IPv4 or IPv6 or domain name.

 - *log* (type integer) - can be assigned with a value defined in default_decision option or with a value taken from PDP response. The attribute defines the dnstap log level, i.e. the attribute set to be sent with DNStap message. See option **dnstap** above. The valid range is from 0 to 3 (inclusive). The value 0 of log attribute means to not send DNSTAP message at all. The default value for log attribute is 0.

## Example

~~~ txt
policy {
    endpoint 10.0.0.7:1234, 10.0.0.8:1234
    streams 100
    connection_timeout 1s

    edns0 0xffee client_id hex 32 0 32
    edns0 0xffee group_id hex 32 16 32
    edns0 0xffef uid
    edns0 0xffea client_ip ip
    edns0 0xffeb client_name bytes
    validation1 type="query" domain_name
    validation2 type="response" address
    default_decision policy_action=2 log=1 client_ip=192.168.0.101
    metrics client_name
    dnstap 1 policy_action
    dnstap 2 policy_action client_id
    dnstap 3 policy_action client_id group_id uid

    passthrough mycompanyname.com. mycompanyname.org.
    debug_query_suffix debug.
    debug_id instance_1
    log
}
~~~

In example above the edns0 options with code 0xffee is splitted into two attributes - `client_id` (first 16 bytes) and `group_id` (last 16 bytes). Entire option should have size 32 bytes otherwise `client_id` and `group_id` will not be parsed.
