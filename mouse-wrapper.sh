#!/bin/bash
/logger/main --role mouse --logger-pipe /dev/shm &
sleep 3
/collector/collector --role mouse --ip 0.0.0.0 --port 8123 --logger-pipe /dev/shm --input-devices /collector/devices &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?