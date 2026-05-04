#!/bin/bash
# Records argv, env, and stdin to files in $SPY_OUT, then emits a clean
# turn so the SDK is happy.
set -u
out="${SPY_OUT:?SPY_OUT must be set by the test}"
printf '%s\n' "$@" > "$out/argv"
env | sort > "$out/env"
cat > "$out/stdin"
echo '{"type":"thread.started","thread_id":"spy-id"}'
echo '{"type":"turn.completed","usage":{"input_tokens":0,"cached_input_tokens":0,"output_tokens":0,"reasoning_output_tokens":0}}'
exit 0
