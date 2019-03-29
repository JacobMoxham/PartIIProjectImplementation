#!/usr/bin/env bash
docker login

docker build -t jacobmoxham2/part_ii_project_implementation:im-client ./client
sudo docker push jacobmoxham2/part_ii_project_implementation:im-client

docker build -t jacobmoxham2/part_ii_project_implementation:im-server ./server
sudo docker push jacobmoxham2/part_ii_project_implementation:im-server

