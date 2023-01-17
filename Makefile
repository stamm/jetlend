.DEFAULT_GOAL := run
.PHONY: run
run:
	go run -race cmd/cli/main.go

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
upload:
	scp jetlend_bot root@kube1.zagirov.name:/tmp
	ssh root@kube1.zagirov.name 'systemctl stop jetlend-bot.service'
	ssh root@kube1.zagirov.name 'cp /tmp/jetlend_bot /root/jetlend_bot'
	ssh root@kube1.zagirov.name 'systemctl start jetlend-bot.service'
