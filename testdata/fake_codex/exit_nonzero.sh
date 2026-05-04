#!/bin/bash
cat > /dev/null
echo '{"type":"thread.started","thread_id":"01HMFAKE3"}'
echo "boom!" >&2
exit 7
