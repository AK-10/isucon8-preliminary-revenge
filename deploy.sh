cd /home/isucon/torb/webapp/go
make
sudo systemctl restart torb.go.service
cd /home/isucon/torb/bench
./bin/bench -output=./result.json
jq < ./result.json

cd /home/isucon/torb
