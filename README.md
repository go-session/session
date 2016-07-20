Session
=======

> A branch is based on [beego session](http://beego.me/docs/module/session.md)

[![GoDoc](https://godoc.org/gopkg.in/session.v1?status.svg)](https://godoc.org/gopkg.in/session.v1)

Session is a Go session manager. It can use many session providers. Just like the `database/sql` and `database/sql/driver`.

Get
----

``` bash
$ go get -u gopkg.in/session.v1
```

Usage
-----

``` go
var globalSessions *session.Manager
```

### Memory Store

``` go
func init() {
    globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
    go globalSessions.GC()
}
```

### File Store

``` go
func init() {
    globalSessions, _ = session.NewManager("file",`{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"./tmp"}`)
    go globalSessions.GC()
}
```

### Cookie Store

``` go
func init() {
    globalSessions, _ = session.NewManager(
        "cookie", `{"cookieName":"gosessionid","enableSetCookie":false,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`)
    go globalSessions.GC()
}
```

License
-------

```
Copyright (c) 2016, Session
All rights reserved.
```