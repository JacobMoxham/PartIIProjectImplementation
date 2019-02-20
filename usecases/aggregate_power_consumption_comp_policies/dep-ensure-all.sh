#!/usr/bin/env bash
cd server/
dep ensure -update
cd ../requesting-client/
dep ensure -update
cd ../data-client-both/
dep ensure -update
cd  ../data-client-can-compute/
dep ensure -update
cd ../data-client-no-computation/
dep ensure -update
cd ../data-client-raw-data/
dep ensure -update
cd ../data-client-configurable/
dep ensure -update