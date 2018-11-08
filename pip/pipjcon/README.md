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
- **-max-connections** - limit on number of simultaneous connections (defailt - no limit);
- **-buffer-size** - input/output buffer size (default 1MB);
- **-max-message** - limit on single request/response size (default 10kB);
- **-write-interval** - interval to wait for responses if output buffer isn't full (default 50µs).

## JSON Content format and updates

JSON Content format is exactly the same local content as described in root README.md file. Updates can be loaded similarly to that of PDP server with only exception that JCon server accepts only content updates and raises an error in case of policy update.
