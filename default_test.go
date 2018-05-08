package session

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefault(t *testing.T) {
	cookieName := "test_default"
	InitManager(
		SetCookieName(cookieName),
	)

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

	Convey("Test default session", t, func() {
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
