#!/usr/bin/env bash
docker login

docker build -t jacobmoxham2/part_ii_project_implementation:im-client-no-mware ./client
sudo docker push jacobmoxham2/part_ii_project_implementation:im-client-no-mware

docker build -t jacobmoxham2/part_ii_project_implementation:im-server-no-mware ./server
sudo docker push jacobmoxham2/part_ii_project_implementation:im-server-no-mware

