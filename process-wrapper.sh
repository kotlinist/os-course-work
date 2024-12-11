#!/bin/bash
/logger/main --role process --logger-pipe /dev/shm &
sleep 3
/collector/collector --role process --ip 0.0.0.0 --port 8123 --logger-pipe /dev/shm &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?