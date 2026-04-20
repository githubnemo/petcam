#!/bin/sh

SCRIPT_DIR="$(dirname "$(readlink -e "$0")")"

. /home/pi/envs/petcam/bin/activate
python "${SCRIPT_DIR}/app.py"
