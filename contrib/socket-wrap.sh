#!/bin/bash
set -euo pipefail
exec socat STDIO "UNIX-CONNECT:$BORGSOCKETWRAP_SOCKET_PATH"
