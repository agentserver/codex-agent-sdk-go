#!/bin/bash
# Reads prompt on stdin, emits a complete clean turn on stdout.
cat > /dev/null
echo '{"type":"thread.started","thread_id":"01HMFAKE"}'
echo '{"type":"turn.started"}'
echo '{"type":"item.completed","item":{"id":"i1","type":"agent_message","text":"hello"}}'
echo '{"type":"turn.completed","usage":{"input_tokens":1,"cached_input_tokens":0,"output_tokens":2,"reasoning_output_tokens":0}}'
exit 0
