#!/usr/bin/env bash

name="admin"
pass=$(echo -n 'admin' | md5sum | cut -d ' ' -f 1)
group="superuser"

mongo $* <<EOF

use clustermgr

db.createCollection('users');
db.createCollection('groups');
db.createCollection('clusters');
db.createCollection('nodes');
db.createCollection('processes');

db.users.insert({name: 'admin', pass: '21232f297a57a5a743894a0e4a801fc3', group: 'superuser', enable: true{)
db.group.insert({name: 'superuser', authorities: {cluster:7, groupadmin:7, node:7, process:7, task:7, useradmin:7}})

EOF
