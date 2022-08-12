#export IS_TEST_NET=1
prod_name="qq_bot"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $prod_name .
chmod +x $prod_name
kill -9 $(ps -ax | grep $prod_name | grep -v grep | awk '{print $1}')
nohup ./$prod_name 2 >>log >&1 &
#kill $(ps -ax | grep server| grep go | grep -v grep | awk '{print $1}')
#
#nohup go run server.go >> log 2 >&1 &
