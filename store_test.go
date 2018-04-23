package session

import (
	"os"
	"testing"
	"time"
)

func testStore(t *testing.T, mstore ManagerStore) {
	store, err := mstore.Create("store", 2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if store.SessionID() != "store" {
		t.Error("Wrong value obtained")
		return
	}

	store.Set("foo", "bar")
	store.Set("user", "bar")
	err = store.Save()
	if err != nil {
		t.Error(err.Error())
		return
	}

	foo, ok := store.Get("foo")
	if !ok || foo != "bar" {
		t.Error("Wrong value obtained")
		return
	}

	foo = store.Delete("foo")
	if foo != "bar" {
		t.Error("Wrong value obtained")
		return
	}

	err = store.Save()
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, ok = store.Get("foo")
	if ok {
		t.Error("Expected value is false")
		return
	}

	user, ok := store.Get("user")
	if !ok || user != "bar" {
		t.Error("Wrong value obtained")
		return
	}

	store.Flush()
	err = store.Save()
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, ok = store.Get("user")
	if ok {
		t.Error("Expected value is false")
		return
	}
}

func TestMemoryStore(t *testing.T) {
	mstore := NewMemoryStore()
	testStore(t, mstore)
}

func TestFileStore(t *testing.T) {
	mstore := NewFileStore("test.db")
	defer os.Remove("test.db")
	defer mstore.Close()
	testStore(t, mstore)
}

func testManagerStore(t *testing.T, mstore ManagerStore) {
	sid := "manager"
	store, err := mstore.Create(sid, 2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	store.Set("foo", "bar")
	err = store.Save()
	if err != nil {
		t.Error(err.Error())
		return
	}

	store, err = mstore.Update(sid, 2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	foo, ok := store.Get("foo")
	if !ok || foo != "bar" {
		t.Error("Wrong value obtained")
		return
	}

	err = mstore.Delete(sid)
	if err != nil {
		t.Error(err.Error())
		return
	}

	exists, err := mstore.Check(sid)
	if err != nil {
		t.Error(err.Error())
		return
	} else if exists {
		t.Error("Expected value is false")
	}
}

func TestManagerMemoryStore(t *testing.T) {
	mstore := NewMemoryStore()
	defer mstore.Close()
	testManagerStore(t, mstore)
}

func TestManagerFileStore(t *testing.T) {
	mstore := NewFileStore("test_manager.db")
	defer os.Remove("test_manager.db")
	defer mstore.Close()
	testManagerStore(t, mstore)
}

func TestStoreWithExpired(t *testing.T) {
	mstore := NewMemoryStore()

	sid := "test_store_expired"
	store, err := mstore.Create(sid, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}

	store.Set("foo", "bar")
	err = store.Save()
	if err != nil {
		t.Error(err.Error())
		return
	}

	foo, ok := store.Get("foo")
	if !ok || foo != "bar" {
		t.Error("Wrong value obtained")
		return
	}

	time.Sleep(time.Second * 2)

	exists, err := mstore.Check(sid)
	if err != nil {
		t.Error(err.Error())
		return
	} else if exists {
		t.Error("Expected value is false")
	}
}
