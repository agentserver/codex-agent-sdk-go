#!/bin/bash
cat > /dev/null
echo '{"type":"thread.started","thread_id":"01HMFAKE4"}'
# Trap SIGTERM so we can verify SIGKILL escalation.
trap 'echo "ignoring TERM" >&2' TERM
sleep 30
