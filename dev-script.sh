docker run -it --rm \
    -p 8080:8080 \
    -v ./data:/data \
    -e SQLITE_DATABASE="app.db" \
    coleifer/sqlite-web

#  helm upgrade  kubesentry -n operation    ./  -f values.yaml  -f ./env/stag/values.yaml 
# apk add stress-ng && stress-ng --vm 1 --vm-bytes 800M -t 3000000s