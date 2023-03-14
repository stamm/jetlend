This project is gather personal stats on one page stat from crowdlending platform Jetlend.

To run you need:
* install go >= 1.20
* set up env variable `JETLEND_COOKIE`, copy it from browser cookie: sessionid


`make run` - print all stats with quantiles
`make expect` - print expected cashflow (default 7 days, can tuned with env JETLEND_DAYS)


Also you can setup a telegram bot, in this case you need a server. Bot will daily send stat to your TG_USER_ID
Set up a systemd.unit in /etc/systemd/system/jetlend-bot.service:
```
[Unit]
Description=Jetlend bot

[Service]
ExecStart=/root/jetlend_bot

[Install]
WantedBy=multi-user.target
```

Set tokens and cookies (/etc/systemd/system/jetlend-bot.service.d/10-env.conf):
```
[Service]
Environment=JETLEND_CFG="{TG_USER_ID}={COOKIE}"
Environment="TELEGRAM_APITOKEN={TOKEN}"
```

