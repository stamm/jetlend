This project is gather personal stats on one page stat from crowdlending platform Jetlend.

To run you need:
* install go >= 1.20
* set up env variable `JETLEND_COOKIE`, copy it from browser cookie: sessionid


`make run` - print all stats with quantiles
`make expect` - print expected cashflow (default 7 days, can tuned with env JETLEND_DAYS)
