#!/usr/bin/env bash

name="admin"
pass=$(echo -n 'admin' | md5sum | cut -d ' ' -f 1)
group="superuser"

mongo $* <<EOF
