#!/bin/bash
cat > /dev/null
echo '{"type":"thread.started","thread_id":"01HMFAKE2"}'
echo '{"type":"turn.failed","error":{"message":"model rejected"}}'
exit 0
