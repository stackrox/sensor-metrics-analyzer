#!/bin/sh
set -e

if [ -d /dev/shm ]; then
  export TMPDIR=/dev/shm
  mkdir -p \
    /dev/shm/nginx/client_body \
    /dev/shm/nginx/proxy \
    /dev/shm/nginx/fastcgi \
    /dev/shm/nginx/uwsgi \
    /dev/shm/nginx/scgi
else
  mkdir -p \
    /tmp/nginx/client_body \
    /tmp/nginx/proxy \
    /tmp/nginx/fastcgi \
    /tmp/nginx/uwsgi \
    /tmp/nginx/scgi
fi

if [ -z "${LISTEN_ADDR:-}" ]; then
  export LISTEN_ADDR=":8081"
fi

/app/web-server &

exec nginx -e /dev/stderr -g 'daemon off;' -c /etc/nginx/nginx.conf
