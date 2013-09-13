trap 'kill 0' SIGINT SIGTERM EXIT

LOG_BIN=./main
MACHINES="127.0.0.1:7770,127.0.0.1:7771,127.0.0.1:7772,127.0.0.1:7773"
LOG_PREFIX="/tmp/"

echo -e "123|hello\n456|world\n" > ${LOG_PREFIX}machine.0.log
echo -e "123|hello\n3249845646|system down\n47564948|clocks synced\n" > ${LOG_PREFIX}machine.1.log
echo -e "123|hello\n24524555|no more pizza\n4668254452|pizza restocked\n367853844|ice cream melted" > ${LOG_PREFIX}machine.2.log
printf '333|repeat%s\n' {1..100000} > ${LOG_PREFIX}machine.3.log


$LOG_BIN --bind=":7770" -machines=$MACHINES -logs="${LOG_PREFIX}machine.0.log" -batch &
sleep 0.1 && echo
$LOG_BIN --bind=":7771" -machines=$MACHINES -logs="${LOG_PREFIX}machine.1.log" -batch &
sleep 0.1 && echo
$LOG_BIN --bind=":7772" -machines=$MACHINES -logs="${LOG_PREFIX}machine.2.log" -batch &
sleep 0.1 && echo
$LOG_BIN --bind=":7773" -machines=$MACHINES -logs="${LOG_PREFIX}machine.3.log"
