#!/bin/bash
set -e
curl "http://ocrproxy-myproject.192.168.178.31.nip.io/v2/myproject/$1/blobs/sha256:$2" > temp/$1/$2.tar.gz
mkdir -p temp/$1/$2
tar xzf temp/$1/$2.tar.gz -C temp/$1/$2
rm temp/$1/$2.tar.gz
