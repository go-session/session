package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSessionStart(t *testing.T) {
	cookieName := "test_session_start"
	manager := NewManager(
		SetCookieName(cookieName),
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := manager.Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("login") == "1" {
			foo, ok := store.Get("foo")
			fmt.Fprintf(w, "%v:%v", foo, ok)
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

	Convey("Test session start", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, cookieName)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s?login=1", ts.URL), nil)
		So(err, ShouldBeNil)
		req.AddCookie(cookie)

		res, err = http.DefaultClient.Do(req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)

		buf, err := ioutil.ReadAll(res.Body)
		So(err, ShouldBeNil)
		res.Body.Close()
		So(string(buf), ShouldEqual, "bar:true")
	})
}

func TestSessionDestroy(t *testing.T) {
	cookieName := "test_session_destroy"

	manager := NewManager(
		SetCookieName(cookieName),
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("logout") == "1" {
			err := manager.Destroy(nil, w, r)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Fprint(w, "ok")
			return
		}

		store, err := manager.Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("check") == "1" {
			foo, ok := store.Get("foo")
			fmt.Fprintf(w, "%v:%v", foo, ok)
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

	Convey("Test session destroy", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, cookieName)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s?logout=1", ts.URL), nil)
		So(err, ShouldBeNil)

		req.AddCookie(cookie)
		res, err = http.DefaultClient.Do(req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)

		req, err = http.NewRequest("GET", fmt.Sprintf("%s?check=1", ts.URL), nil)
		So(err, ShouldBeNil)
		req.AddCookie(cookie)
		res, err = http.DefaultClient.Do(req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)

		buf, err := ioutil.ReadAll(res.Body)
		So(err, ShouldBeNil)
		res.Body.Close()
		So(string(buf), ShouldEqual, "<nil>:false")
	})
}

func TestSessionRefresh(t *testing.T) {
	cookieName := "test_session_refresh"

	manager := NewManager(
		SetCookieName(cookieName),
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := manager.Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("refresh") == "1" {
			vstore, verr := manager.Refresh(nil, w, r)
			if verr != nil {
				t.Error(err)
				return
			}

			if vstore.SessionID() == store.SessionID() {
				t.Errorf("Not expected value")
				return
			}

			foo, ok := vstore.Get("foo")
			fmt.Fprintf(w, "%s:%v", foo, ok)
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

	Convey("Test session refresh", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, cookieName)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s?refresh=1", ts.URL), nil)
		So(err, ShouldBeNil)

		req.AddCookie(cookie)
		res, err = http.DefaultClient.Do(req)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)
		So(res.Cookies()[0].Value, ShouldNotEqual, cookie.Value)

		buf, err := ioutil.ReadAll(res.Body)
		So(err, ShouldBeNil)
		res.Body.Close()
		So(string(buf), ShouldEqual, "bar:true")
	})
}
