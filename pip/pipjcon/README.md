# PIPJCon (Policy Information Point JSON Content Server)

PIPJCon is a server for JSON Content over network or unix socket. The server acts similarly to PDP's local selector but remotely.

## Usage

Usage of pipjcon:
```
$ pipjcon [-network <network>] [-a <service-address>] [-c <control-address>] [-j <jcon-file>] [...]
```
Options:
- **-network** - type of network to listen at (default "tcp");
- **-a** - address to listen at (default for "tcp\*" - localhost:5600, default for "unix" - /var/run/pip.socket);
- **-c** - address for control (default for "tcp\*" - localhost:5604, unavailable for "unix");
- **-j** - path to JCon file to load at startup;
- **-w** - number of workers per connection (default 100);
- **-max-connections** - limit on number of simultaneous connections (defailt - no limit);
- **-buffer-size** - input/output buffer size (default 1MB);
- **-max-message** - limit on single request/response size (default 10kB);
- **-max-args** - limit on number of arguments for a request (default 32);
- **-write-interval** - interval to wait for responses if output buffer isn't full (default 50µs).

## JSON Content format and updates

JSON Content format is exactly the same local content as described in root README.md file. Updates can be loaded similarly to that of PDP server with only exception that JCon server accepts only content updates and raises an error in case of policy update. For example content from file examples/07-selector/content.json can be read by server at startup:
```
$ pipjcon -j content.json
INFO[0000] PIP JCon server                              
INFO[0000] opening content                               content=content.json
INFO[0000] parsing content                               content=content.json
INFO[0000] opening service port                          address="localhost:5600" network=tcp
```

Or it can be uploaded by papcli:
```
# First terminal
$ papcli -s localhost:5602 -j content.json -vt 823f79f2-0001-4eb2-9ba0-2a8c1b284443
INFO[0000] Requesting data upload to PDP servers...     
INFO[0000] Uploading data to PDP servers...             
```

PIPJCon working in the second terminal prints following output in the case:
```
# Second terminal
$ pipjcon -c localhost:5602
INFO[0000] PIP JCon server                              
INFO[0000] opening control port                          address="localhost:5602"
INFO[0005] control request                              
INFO[0005] request has been registered                   req-id=1
INFO[0005] data stream                                  
INFO[0005] uploading data for request                    req-id=1
INFO[0005] stream has been read and parsed as snapshot   size=477
INFO[0005] apply command                                 req-id=1
INFO[0005] new content has been applied                  ctn-id= tag=823f79f2-0001-4eb2-9ba0-2a8c1b284443
INFO[0005] opening service port                          address="localhost:5600" network=tcp
```

Then an update looks like this:
```
# First terminal
$ papcli -s localhost:5602 -id content -j update.json -vf 823f79f2-0001-4eb2-9ba0-2a8c1b284443 -vt 93a17ce2-788d-476f-bd11-a5580a2f35f3
INFO[0000] Requesting data upload to PDP servers...     
INFO[0000] Uploading data to PDP servers...             
```

PIPJCon reacts with following logs to the update:
```
# Second terminal
...
INFO[0010] control request                              
INFO[0010] request has been registered                   req-id=2
INFO[0010] data stream                                  
INFO[0010] uploading data for request                    req-id=2
INFO[0010] stream has been read and parsed as update     size=555
INFO[0010] apply command                                 req-id=2
INFO[0010] content update has been applied               ctn-id=content curr-tag=93a17ce2-788d-476f-bd11-a5580a2f35f3 prev-tag=823f79f2-0001-4eb2-9ba0-2a8c1b284443 req-id=2
```

Content of update.json file can be following:
```json
[
  {
    "op": "delete",
    "path": ["domain-addresses", "good", "example.com"]
  },
  {
    "op": "add",
    "path": ["domain-addresses", "good", "example.com"],
    "entity": {
      "type": "set of networks",
      "data": ["2001:db8:1000::/40", "2001:db8:2000::/40"]
    }
  },
  {
    "op": "delete",
    "path": ["domain-addresses", "bad", "example.com"]
  },
  {
    "op": "add",
    "path": ["domain-addresses", "bad", "example.com"],
    "entity": {
      "type": "set of networks",
      "data": ["192.0.2.16/28", "192.0.2.32/28"]
    }
  }
]

```
