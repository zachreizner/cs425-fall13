trap 'kill 0' SIGINT SIGTERM EXIT

LOG_BIN='./main'
MACHINES="linux3.ews.illinois.edu:8886,linux4.ews.illinois.edu:7771,linux5.ews.illinois.edu:7772,linux6.ews.illinois.edu:7773"
LOG_PREFIX=""
NETID=

ssh ${NETID}@linux3.ews.illinois.edu "$LOG_BIN --bind=\":8886\" -machines=$MACHINES -logs=\"${LOG_PREFIX}machine.0.log\" -batch" &
sleep 0.5 && echo
ssh ${NETID}@linux4.ews.illinois.edu "$LOG_BIN --bind=\":7771\" -machines=$MACHINES -logs=\"${LOG_PREFIX}machine.1.log\" -batch" &
sleep 0.5 && echo
ssh ${NETID}@linux5.ews.illinois.edu "$LOG_BIN --bind=\":7772\" -machines=$MACHINES -logs=\"${LOG_PREFIX}machine.2.log\" -batch" &
sleep 0.5 && echo
ssh ${NETID}@linux6.ews.illinois.edu "$LOG_BIN --bind=\":7773\" -machines=$MACHINES -logs=\"${LOG_PREFIX}machine.3.log\""
