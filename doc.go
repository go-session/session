// Package session implements a efficient, safely and easy-to-use session library for Go.

/*
Example:

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

Learn more at https://github.com/go-session/session
*/

package session
