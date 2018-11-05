# MkPIPHandler example

Here placed a simple example of PIP handler schema and generated output.

## Schema

The example contains schema of a package with name "pipexample". It defines handler accepting requests with string and address arguments. A response should be a network. See schema.yaml.

## Package pipexample

Resulting package placed in "pipexample" subdirectory. It's obtained with command:
```
$ mkpiphandler -s schema.yaml -d .
INFO[0000] making pip handler                            output=. schema=schema.yaml
```

## Run example server

You can try a PIP server with the handler:
```
$ go run server.go
INFO[0000] PIP example server
INFO[0000] Binding server
INFO[0000] Serving requests
```
