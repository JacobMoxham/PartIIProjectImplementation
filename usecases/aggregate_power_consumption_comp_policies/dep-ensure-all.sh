#!/usr/bin/env bash
cd server/
dep ensure -update
cd ../requesting-client/
dep ensure -update
cd ../data-client-configurable/
dep ensure -update