#!/usr/bin/env bash
docker login

docker build -t jacobmoxham2/part_ii_project_implementation:data-configurable ./data-client-configurable
sudo docker push jacobmoxham2/part_ii_project_implementation:data-configurable

docker build -t jacobmoxham2/part_ii_project_implementation:requesting ./requesting-client
sudo docker push jacobmoxham2/part_ii_project_implementation:requesting

docker build -t jacobmoxham2/part_ii_project_implementation:server ./server
sudo docker push jacobmoxham2/part_ii_project_implementation:server

