package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var defaultCookieName = "test_default_start"

func init() {
	InitManager(
		SetCookieName(defaultCookieName),
	)
}

func TestDefaultStart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := Start(nil, w, r)
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

	Convey("Test default start", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, defaultCookieName)

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

func TestDefaultDestroy(t *testing.T) {
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

	Convey("Test default destroy", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, defaultCookieName)

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

func TestDefaultRefresh(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := Start(nil, w, r)
		if err != nil {
			t.Error(err)
			return
		}

		if r.URL.Query().Get("refresh") == "1" {
			vstore, verr := Refresh(nil, w, r)
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

	Convey("Test default refresh", t, func() {
		res, err := http.Get(ts.URL)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(len(res.Cookies()), ShouldBeGreaterThan, 0)

		cookie := res.Cookies()[0]
		So(cookie.Name, ShouldEqual, defaultCookieName)

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
