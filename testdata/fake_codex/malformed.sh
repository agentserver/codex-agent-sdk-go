#!/bin/bash
cat > /dev/null
echo '{"type":"thread.started","thread_id":"01HMFAKE5"}'
echo '{not valid json'
echo '{"type":"turn.completed","usage":{"input_tokens":0,"cached_input_tokens":0,"output_tokens":0,"reasoning_output_tokens":0}}'
exit 0
