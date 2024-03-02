.DEFAULT_GOAL := run
MODE ?= stat
RACE ?= -race
.PHONY: run
run:
	go run ${RACE} cmd/cli/main.go -m ${MODE}
run_alt:
	go run ${RACE} -tags alt cmd/cli/main.go -m ${MODE}
expect:
	go run ${RACE} cmd/cli/main.go -m expect
csv:
	go run ${RACE} cmd/csv/main.go > intelinvest_test1.csv
action:
	go run ${RACE} cmd/cli/main.go -m what_buy
secondary:
	go run ${RACE} cmd/cli/main.go -m secondary

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
