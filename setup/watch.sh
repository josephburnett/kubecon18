#!/bin/bash

clear
echo "watching for changes to app/service.yaml"

while inotifywait -q -e close_write app/service.yaml; do clear ; kubectl apply -f app/service.yaml; done

