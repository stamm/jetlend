.DEFAULT_GOAL := run
MODE ?= stat
.PHONY: run
run:
	go run -race cmd/cli/main.go -m ${MODE}
expect:
	go run -race cmd/cli/main.go -m expect
csv:
	go run -race cmd/csv/main.go > intelinvest_test1.csv

.PHONY: bot
bot:
	go run -race cmd/bot/main.go

.PHONY: test
test:
	go test -race ./...

.PHONY: bot_linux
bot_linux:
	env GOOS=linux GOARCH=amd64 go build -o jetlend_bot ./cmd/bot/main.go

.PHONY: upload
upload: bot_linux
	scp jetlend_bot ${JETLEND_SSH}:/tmp
	ssh ${JETLEND_SSH} 'systemctl stop jetlend-bot.service'
	ssh ${JETLEND_SSH} 'cp /tmp/jetlend_bot /root/jetlend_bot'
	ssh ${JETLEND_SSH} 'systemctl start jetlend-bot.service'
