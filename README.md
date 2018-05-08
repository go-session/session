# session

> A efficient, safely and easy-to-use session library for Go. 

[![Build][Build-Status-Image]][Build-Status-Url] [![Coverage][Coverage-Image]][Coverage-Url] [![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

## Quick Start

### Download and install

```bash
$ go get -u -v gopkg.in/session.v2
```

### Create file `server.go`

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/session.v2"
)

func main() {
	session.InitManager(
		session.SetCookieName("session_id"),
		session.SetSign([]byte("sign")),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		store.Set("foo", "bar")
		err = store.Save()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		http.Redirect(w, r, "/foo", 302)
	})

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		foo, ok := store.Get("foo")
		if ok {
			fmt.Fprintf(w, "foo:%s", foo)
			return
		}
		fmt.Fprint(w, "does not exist")
	})

	http.ListenAndServe(":8080", nil)
}
```

### Build and run

```bash
$ go build server.go
$ ./server
```

### Open in your web browser

<http://localhost:8080>

    foo:bar

## Features

- Easy to use
- Multi-storage support
- More secure, signature-based tamper-proof
- Context support

## Store Implementations

- [Memory Store](https://github.com/go-session/session/blob/master/store.go#L50) - [buntdb](https://github.com/tidwall/buntdb)
- [File Store](https://github.com/go-session/session/blob/master/store.go#L60) - [buntdb](https://github.com/tidwall/buntdb)
- [https://github.com/go-session/redis](https://github.com/go-session/redis) - Redis
- [https://github.com/go-session/cookie](https://github.com/go-session/cookie) - Cookie

## MIT License

    Copyright (c) 2018 Lyric

[reportcard-url]: https://goreportcard.com/report/gopkg.in/session.v2
[reportcard-image]: https://goreportcard.com/badge/gopkg.in/session.v2
[Build-Status-Url]: https://travis-ci.org/go-session/session
[Build-Status-Image]: https://travis-ci.org/go-session/session.svg?branch=master
[Coverage-Url]: https://coveralls.io/github/go-session/session?branch=master
[Coverage-Image]: https://coveralls.io/repos/github/go-session/session/badge.svg?branch=master
[godoc-url]: https://godoc.org/gopkg.in/session.v2
[godoc-image]: https://godoc.org/gopkg.in/session.v2?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg
