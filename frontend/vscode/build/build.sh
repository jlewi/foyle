#!/bin/bash
# Script to build vscode
set -ex

git clone https://github.com/microsoft/vscode.git /vscode

cd /vscode
git checkout 1.85.0

# Call yarn to initialize the yarn project
yarn
yarn compile-build
yarn compile-extensions-build
node --max_old_space_size=4095 ./node_modules/gulp/bin/gulp.js compile-extension-media
yarn minify-vscode-reh-web