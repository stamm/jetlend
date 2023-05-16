This project is gather personal stats on one page stat from crowdlending platform Jetlend.

To run you need:
* install go >= 1.20
* set up env variable `JETLEND_COOKIE`, copy it from browser cookie: sessionid


`make run` - print all stats with quantiles

`make expect` - print expected cashflow (default 7 days, can tuned with env `JETLEND_DAYS`)

`make csv` - export transaction to IntelInvest's format. Need to [download all transaction](https://jetlend.ru/invest/v3/notifications), convert from xslx to csv (file transactions-2.csv)

default values:

```
next target: 0.1%
target: 0.2%
max target: 0.3%
max sum: 6_000

if > 372 days
    max sum: 6_000 / 4 = 1_500 (0.05%)
if > 30% days
    max sum: 6_000 / 6 = 1_000 (0.033%)
max sum: 6_000 / 2 = 3_000  (0.1%)
```


Also you can setup a telegram bot, in this case you need a server. Bot will daily send stat to your `TG_USER_ID`
Set up a systemd unit in `/etc/systemd/system/jetlend-bot.service`:
```
[Unit]
Description=Jetlend bot

[Service]
ExecStart=/root/jetlend_bot

[Install]
WantedBy=multi-user.target
```

Set tokens and cookies `/etc/systemd/system/jetlend-bot.service.d/10-env.conf`:
```
[Service]
Environment=JETLEND_CFG="{TG_USER_ID}={COOKIE}"
Environment="TELEGRAM_APITOKEN={TOKEN}"
```

Start the unit:
`systemctl daemon-reload && systemctl restart jetlend-bot.service`

In case of debug, you can setup schedule sending:
```
Environment=JETLEND_SCHEDULE="04 8 * * *"
```
