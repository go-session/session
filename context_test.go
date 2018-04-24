package session

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctxKey interface{} = "ctx"
		ctx := context.WithValue(context.Background(), ctxKey, "bar")
		store, err := Start(ctx, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		ctxValue := store.Context().Value(ctxKey)
		if !reflect.DeepEqual(ctxValue, "bar") {
			t.Error("Not expected value:", ctxValue)
			return
		}

		req, ok := FromReqContext(store.Context())
		if !ok || req.URL.Query().Get("foo") != "bar" {
			t.Error("Not expected value:", req.URL.Query().Get("foo"))
			return
		}

		res, ok := FromResContext(store.Context())
		if !ok {
			t.Error("Not expected value")
			return
		}

		fmt.Fprint(res, "ok")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL + "?foo=bar")
	if err != nil {
		t.Error(err)
		return
	}

	buf, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if string(buf) != "ok" {
		t.Error("Not expected value:", string(buf))
	}
}
