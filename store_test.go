package session

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func testStore(store Store) {
	foo, ok := store.Get("foo")
	So(ok, ShouldBeFalse)
	So(foo, ShouldBeNil)

	store.Set("foo", "bar")
	store.Set("foo2", "bar2")
	err := store.Save()
	So(err, ShouldBeNil)

	foo, ok = store.Get("foo")
	So(ok, ShouldBeTrue)
	So(foo, ShouldEqual, "bar")

	foo = store.Delete("foo")
	So(foo, ShouldEqual, "bar")

	foo, ok = store.Get("foo")
	So(ok, ShouldBeFalse)
	So(foo, ShouldBeNil)

	foo2, ok := store.Get("foo2")
	So(ok, ShouldBeTrue)
	So(foo2, ShouldEqual, "bar2")

	err = store.Flush()
	So(err, ShouldBeNil)

	foo2, ok = store.Get("foo2")
	So(ok, ShouldBeFalse)
	So(foo2, ShouldBeNil)
}

func TestMemoryStore(t *testing.T) {
	mstore := NewMemoryStore()
	defer mstore.Close()

	Convey("Test memory storage operation", t, func() {
		store, err := mstore.Create(context.Background(), "test_memory_store", 10)
		if err != nil {
			So(err, ShouldBeNil)
		}
		testStore(store)
	})
}

func TestFileStore(t *testing.T) {
	mstore := NewFileStore("test.db")
	defer os.Remove("test.db")
	defer mstore.Close()

	Convey("Test file storage operation", t, func() {
		store, err := mstore.Create(context.Background(), "test_file_store", 10)
		if err != nil {
			So(err, ShouldBeNil)
		}
		testStore(store)
	})
}

func testManagerStore(mstore ManagerStore) {
	sid := "test_manager_store"
	store, err := mstore.Create(context.Background(), sid, 10)
	So(store, ShouldNotBeNil)
	So(err, ShouldBeNil)

	store.Set("foo", "bar")
	err = store.Save()
	So(err, ShouldBeNil)

	store, err = mstore.Update(context.Background(), sid, 10)
	So(store, ShouldNotBeNil)
	So(err, ShouldBeNil)

	foo, ok := store.Get("foo")
	So(ok, ShouldBeTrue)
	So(foo, ShouldEqual, "bar")

	newsid := "test_manager_store2"
	store, err = mstore.Refresh(context.Background(), sid, newsid, 10)
	So(store, ShouldNotBeNil)
	So(err, ShouldBeNil)

	foo, ok = store.Get("foo")
	So(ok, ShouldBeTrue)
	So(foo, ShouldEqual, "bar")

	exists, err := mstore.Check(context.Background(), sid)
	So(exists, ShouldBeFalse)
	So(err, ShouldBeNil)

	err = mstore.Delete(context.Background(), newsid)
	So(err, ShouldBeNil)

	exists, err = mstore.Check(context.Background(), newsid)
	So(exists, ShouldBeFalse)
	So(err, ShouldBeNil)
}

func TestManagerMemoryStore(t *testing.T) {
	mstore := NewMemoryStore()
	defer mstore.Close()

	Convey("Test memory-based storage management operations", t, func() {
		testManagerStore(mstore)
	})
}

func TestManagerFileStore(t *testing.T) {
	mstore := NewFileStore("test_manager.db")
	defer os.Remove("test_manager.db")
	defer mstore.Close()
	Convey("Test file-based storage management operations", t, func() {
		testManagerStore(mstore)
	})
}

func testStoreWithExpired(mstore ManagerStore) {
	sid := "test_store_expired"
	store, err := mstore.Create(context.Background(), sid, 1)
	So(store, ShouldNotBeNil)
	So(err, ShouldBeNil)

	store.Set("foo", "bar")
	err = store.Save()
	So(err, ShouldBeNil)

	store, err = mstore.Update(context.Background(), sid, 1)
	So(store, ShouldNotBeNil)
	So(err, ShouldBeNil)

	foo, ok := store.Get("foo")
	So(foo, ShouldEqual, "bar")
	So(ok, ShouldBeTrue)

	time.Sleep(time.Second * 2)

	exists, err := mstore.Check(context.Background(), sid)
	So(err, ShouldBeNil)
	So(exists, ShouldBeFalse)
}

func TestMemoryStoreWithExpired(t *testing.T) {
	mstore := NewMemoryStore()
	defer mstore.Close()

	Convey("Test Memory Store Expiration", t, func() {
		testStoreWithExpired(mstore)
	})
}

func TestFileStoreWithExpired(t *testing.T) {
	mstore := NewFileStore("test_expired.db")
	defer os.Remove("test_expired.db")
	defer mstore.Close()

	Convey("Test File Store Expiration", t, func() {
		testStoreWithExpired(mstore)
	})
}
