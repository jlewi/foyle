#!/bin/bash
# Build the frontend.
# This is useful for running in a container to build the fronted
set -ex
yarn
yarn package-web

cat > dist/version.json <<EOF
{
  "version": "$VERSION",
  "date": "$DATE",
  "commit": "$COMMIT"
}
EOF

