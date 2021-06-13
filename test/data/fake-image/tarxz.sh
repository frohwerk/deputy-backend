#!/bin/bash
find * -maxdepth 1 -type f -name '*.tar.gz' -print0                             |\
sed -z 's/\.tar\.gz$//'                                                         |\
xargs -0 sh -c 'for arg do mkdir "$arg"; tar xzf "$arg.tar.gz" -C "$arg"; done' _
