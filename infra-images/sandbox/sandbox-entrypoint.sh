#!/usr/bin/env sh

set -e
set -o pipefail

if [ ! -f /usr/bin/dboxed ]; then
  echo "dboxed binary is missing. this image can only be used by 'dboxed sandbox run'"
  exit 1
fi

function do_log() {
    echo "$(date -Iseconds) - $1" >> /var/lib/dboxed/logs/sandbox.log
}

do_log "sandbox started"

function kill_dockerd() {
  if [ -f /var/run/docker.pid ]; then
    do_log "killing dockerd"
    docker_pid=$(cat /var/run/docker.pid)
    kill $docker_pid
    wait $docker_pid
  fi
}

function do_exit() {
  do_log "handing exit"
  set +e
  docker_ps=$(docker ps --format='{{.Names}}')
  set -e
  if echo "$docker_ps" | grep '^dboxed-run-in-sandbox$' &> /dev/null; then
    do_log "stopping dboxed-run-in-sandbox"
    docker stop dboxed-run-in-sandbox || true
  fi
  if echo "$docker_ps" | grep '^dboxed-run-in-sandbox-status$' &> /dev/null; then
    do_log "stopping dboxed-run-in-sandbox-status"
    docker stop dboxed-run-in-sandbox-status || true
  fi

  kill_dockerd
  exit
}
trap do_exit TERM INT

do_log "starting dockerd"
(
  while [ ! -f /exit-marker ]; do
    dockerd --host unix:///var/run/docker.sock --log-format=json 2>&1 | s6-log n20 s1000000 /var/dboxed/logs/dockerd
  done
) &

do_log "waiting for dockerd"
while ! docker ps &> /dev/null; do
  sleep 1
done

do_log "loading busybox image"
docker load < /busybox-image.tar

do_log "starting netns-holder"
dboxed sandbox netns-holder &

do_log "starting dboxed containers"
docker run -d $ENV_FILE_ARG --restart=on-failure --net=host --pid=host --name=dboxed-dns-proxy \
  -v/:/hostfs \
  --privileged --init \
  dboxed-busybox chroot /hostfs dboxed sandbox dns-proxy
docker run -d $ENV_FILE_ARG --restart=on-failure --net=host --pid=host --name=dboxed-run-in-sandbox-status \
  -v/:/hostfs \
  --privileged --init \
  -eDBOXED_SANDBOX=1 \
  dboxed-busybox chroot /hostfs dboxed sandbox run-in-sandbox-status
docker run -d $ENV_FILE_ARG --restart=on-failure --net=host --pid=host --name=dboxed-run-in-sandbox \
  -v/:/hostfs \
  --privileged --init \
  -eDBOXED_SANDBOX=1 \
  dboxed-busybox chroot /hostfs dboxed sandbox run-in-sandbox

sleep infinity
