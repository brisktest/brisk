# alias go="go1.18beta2"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/worker.bin server/main.go &
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/super.bin super/main.go & 

# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a  -o build/cli.bin  brisk-cli/main.go & 

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a  -o build/monitor.bin  monitor/main.go & 
wait
# go build -o build/brisk brisk-cli/main.go &


