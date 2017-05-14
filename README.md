# go-requestfmtlogger

Generate a log line per each request. Adding to current request loggers with go-requestfmtlogger you can add key-value pairs to the log message at any time during the request.

### Usage

Firstly import the package:
`import log "github.com/kostaskoukouvis/go-requestfmtlogger"`

Initialize the logger by specifying in it's config whether it should write to terminal (false)
or the syslog (true):
`logger := log.LoggerConfig{false}`

For usage with the chi router package, use as a normal middleware:
`r.Use(logger.RequestLogger)`

To add log items to the request, import the package in any file you need it and call the Log function
passing it the current request object (r), a human readable message, and anything else you want to add in a comma separated
string - value fashion.
`log.Log(r, "error getting product", "err", err, "product_id", pid)`


### Based and inspired from work such as:
* [log15](https://github.com/inconshreveable/log15)
* [chi-router middleware](https://github.com/pressly/chi)

### Credits:
* [whadron](https://github.com/whadron)
