#!/bin/bash

# Change to the directory with our code that we plan to work from
cd "$GOPATH/src/heartfort"

echo "==== Releasing heartfort ===="
echo "  Deleting the local binary if it exists (so it isn't uploaded)..."
rm heartfort
echo "  Done!"

echo "  Deleting existing code..."
ssh root@123.123.22.33 "rm -rf /root/go/src/heartfort"
echo "  Code deleted successfully!"

echo "  Uploading code..."
rsync -avr --exclude '.git/*' --exclude 'tmp/*' ./ \
  root@123.123.22.33:/root/go/src/heartfort/
echo "  Code uploaded successfully!"

echo "  Go getting deps..."
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get golang.org/x/crypto/bcrypt"
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get github.com/gorilla/mux"
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get github.com/gorilla/schema"
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get github.com/lib/pq"
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get github.com/jinzhu/gorm"
ssh root@123.123.22.33 "export GOPATH=/root/go; \
  /usr/local/go/bin/go get github.com/gorilla/csrf"

echo "  Building the code on remote server..."
ssh root@123.123.22.33 'export GOPATH=/root/go; \
  cd /root/app; \
  /usr/local/go/bin/go build -o ./server \
    $GOPATH/src/heartfort/*.go'
echo "  Code built successfully!"

echo "  Moving assets..."
ssh root@123.123.22.33 "cd /root/app; \
  cp -R /root/go/src/heartfort/assets ."
echo "  Assets moved successfully!"

echo "  Moving views..."
ssh root@123.123.22.33 "cd /root/app; \
  cp -R /root/go/src/heartfort/views ."
echo "  Views moved successfully!"

echo "  Moving Caddyfile..."
ssh root@123.123.22.33 "cd /root/app; \
  cp /root/go/src/heartfort/Caddyfile ."
echo "  Views moved successfully!"

echo "  Restarting the server..."
ssh root@123.123.22.33 "sudo service heartfort restart"
echo "  Server restarted successfully!"

echo "  Restarting Caddy server..."
ssh root@123.123.22.33 "sudo service caddy restart"
echo "  Caddy restarted successfully!"

echo "==== Done releasing heartfort ===="
