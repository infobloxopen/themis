#!/bin/sh

BASENAME=service
PROTOFILE=$BASENAME.proto
SERVICEFILE=$BASENAME.pb.go
OUTPUTDIR=../pip-service

CMD="protoc --go_out=plugins=grpc:$OUTPUTDIR $PROTOFILE"
echo "Generating $OUTPUTDIR/$SERVICEFILE"
$CMD
