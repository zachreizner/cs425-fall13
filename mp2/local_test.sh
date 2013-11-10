trap 'kill 0' SIGINT SIGTERM EXIT

BIN=./main
MACHINES="127.0.0.1:7770,127.0.0.1:7771,127.0.0.1:7772,127.0.0.1:7773"
LOG_PREFIX="/tmp/"

$BIN --bind=":7770" -name="doctor_rhymes" &
sleep 0.1
$BIN --bind=":7771" -seed="127.0.0.1:7770" -name="astral_flight" &
sleep 0.1
$BIN --bind=":7772" -seed="127.0.0.1:7770" -name="njord_gunner"
# $LOG_BIN --bind=":7772" -machines=$MACHINES -logs="${LOG_PREFIX}machine.2.log" -batch &
# sleep 0.1 && echo
# $LOG_BIN --bind=":7773" -machines=$MACHINES -logs="${LOG_PREFIX}machine.3.log"
