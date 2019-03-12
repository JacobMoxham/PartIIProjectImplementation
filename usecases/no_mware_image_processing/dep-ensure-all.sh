#!/usr/bin/env bash
cd server/
dep ensure -update
cd ../client/
dep ensure -update