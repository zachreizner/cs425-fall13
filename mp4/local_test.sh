RUN=true

control_c()
{
    kill 0
    exit 0
}

trap control_c SIGINT EXIT

BIN=./main
MACHINES="127.0.0.1:7770,127.0.0.1:7771,127.0.0.1:7772,127.0.0.1:7773"
LOG_PREFIX="/tmp/"

$BIN --bind=":7770" -name="doctor_rhymes" &
sleep 0.1
$BIN --bind=":7771" -seed="127.0.0.1:7770" -name="astral_flight" &
sleep 0.1
$BIN --bind=":7772" -seed="127.0.0.1:7770" -name="njord_gunner" &
sleep 0.1
$BIN --bind=":7773" -seed="127.0.0.1:7770" -name="ivory_viper" &
while true; do read x; done