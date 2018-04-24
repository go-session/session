package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionStart(t *testing.T) {
	cookieName := "test_session_start"
	InitManager(
		SetCookieName(cookieName),
		SetCookieLifeTime(60),
		SetExpired(10),
		SetSign([]byte("foo")),
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("login") == "1" {
			foo, ok := store.Get("foo")
			if !ok || foo != "bar" {
				t.Error("Wrong value obtained")
				return
			}
			fmt.Fprint(w, "ok")
			return
		}

		store.Set("foo", "bar")
		err = store.Save()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}

	if len(res.Cookies()) == 0 {
		t.Error("No cookie found")
		return
	}

	cookie := res.Cookies()[0]
	if cookie.Name != cookieName {
		t.Error("Not expected value:", cookie.Name)
		return
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?login=1", ts.URL), nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.AddCookie(cookie)

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	res.Body.Close()
	if string(buf) != "ok" {
		t.Error("Not expected value:", string(buf))
		return
	}
}

func TestSessionDestroy(t *testing.T) {
	cookieName := "test_session_destroy"
	InitManager(
		SetCookieName(cookieName),
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("logout") == "1" {
			err := Destroy(nil, w, r)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Fprint(w, "ok")
			return
		}

		store, err := Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("check") == "1" {
			foo, ok := store.Get("foo")
			if ok || foo == "bar" {
				t.Error("Not expected value:", foo)
				return
			}
			fmt.Fprint(w, "ok")
			return
		}

		store.Set("foo", "bar")
		err = store.Save()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
		return
	} else if len(res.Cookies()) == 0 {
		t.Error("No cookie found")
		return
	}

	cookie := res.Cookies()[0]
	if cookie.Name != cookieName {
		t.Error("Not expected value:", cookie.Name)
		return
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?logout=1", ts.URL), nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.AddCookie(cookie)
	http.DefaultClient.Do(req)

	req, err = http.NewRequest("GET", fmt.Sprintf("%s?check=1", ts.URL), nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.AddCookie(cookie)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	res.Body.Close()
	if string(buf) != "ok" {
		t.Error("Not expected value:", string(buf))
		return
	}
}
