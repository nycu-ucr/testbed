#!/bin/bash
make -f Makefile.throughput
mv ./bin/tp_server ./bin/server ; mv ./bin/tp_client ./bin/client
