#!/bin/bash
# Script to prepare the assets
set -ex
# Get the directory of the currently executing script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Install go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -xvf go1.21.5.linux-amd64.tar.gz
mv go /usr/local/
rm go1.21.5.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc

cd ${DIR}
/usr/local/go/bin/go run . --vscode=/vscode --out=/assets

cat > /assets/version.json <<EOF
{
  "version": "$VERSION",
  "date": "$DATE",
  "commit": "$COMMIT"
}
EOF
