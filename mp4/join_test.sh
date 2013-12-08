trap 'kill 0' SIGINT SIGTERM EXIT

BIN=./main

$BIN --bind=":7770" -name="doctor_rhymes" &
$BIN --bind=":7771" -seed="127.0.0.1:7770" -name="jojo"&
$BIN --bind=":7772" -seed="127.0.0.1:7770" -name="bizzaro"&
$BIN --bind=":7773" -seed="127.0.0.1:7770" -name="edwardo"&
for i in {1..100}
do
  ./main -seed "127.0.0.1:7770" -run "insert $i fish tacos"
done
for i in {1..100}
do
  RA=`shuf -i 0-4294967295 -n 1`
  ./main -seed "127.0.0.1:7770" -run "insert $RA i"
done
$BIN --bind=":7774" -seed="127.0.0.1:7770" -name="astral_flight"
