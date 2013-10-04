trap 'kill 0' SIGINT SIGTERM EXIT

BIN=./main
MACHINES="127.0.0.1:7770,127.0.0.1:7771,127.0.0.1:7772,127.0.0.1:7773"
LOG_PREFIX="/tmp/"

$BIN --bind=":7770"  &
sleep 0.1 && echo
$BIN --bind=":7771" -leader="127.0.0.1:38449"
# sleep 0.1 && echo
# $LOG_BIN --bind=":7772" -machines=$MACHINES -logs="${LOG_PREFIX}machine.2.log" -batch &
# sleep 0.1 && echo
# $LOG_BIN --bind=":7773" -machines=$MACHINES -logs="${LOG_PREFIX}machine.3.log"
