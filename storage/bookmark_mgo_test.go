package storage

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/jhh/puka/model"
	"github.com/manyminds/api2go"
)

var (
	storage BookmarkMgoStorage
	oid     string
)

func TestMain(m *testing.M) {
	if os.Getenv("MONGODB_URI") == "" {
		if err := os.Setenv("MONGODB_URI", "mongodb://localhost/test"); err != nil {
			log.Fatal(err)
		}
	}
	var err error
	storage, err = NewBookmarkMgoStorage()
	if err != nil {
		log.Fatal(err)
	}
	loadTestData()
	code := m.Run()
	storage.Close()
	os.Exit(code)
}

func TestMgoInterface(t *testing.T) {
	var bm interface{} = BookmarkMgoStorage{}
	if inty, ok := bm.(BookmarkStorage); !ok {
		t.Fatalf("%v does not implement BookmarkStorage", inty)
	}
}

func TestGetAll(t *testing.T) {
	r, err := storage.GetAll(Query{})
	if err != nil {
		t.Error(err)
	}
	if len(r) != 1000 {
		t.Errorf("len = %d; want: 1000", len(r))
	}
}

func TestGetAllWithTag(t *testing.T) {
	req := api2go.Request{
		QueryParams: map[string][]string{
			"filter[tag]": []string{"go"},
		},
	}
	r, err := storage.GetAll(NewQuery(req))
	if err != nil {
		t.Error(err)
	}
	if len(r) != 370 {
		t.Errorf("len = %d; want: 370", len(r))
	}
}

func TestGetOne(t *testing.T) {
	bm, err := storage.GetOne(oid)
	if err != nil {
		t.Error(err)
	}
	if bm.GetID() != oid {
		t.Errorf("id = %q; want: %q", bm.GetID(), oid)
	}
}

func TestInsert(t *testing.T) {
	bookmark := model.Bookmark{
		Title:       "Test Title",
		URL:         "http://example.com/testing",
		Description: "This is a test bookmark.",
		Timestamp:   time.Now().Round(time.Second), // db round-trip truncates
		Tags:        []string{"test"},
	}

	if err := storage.Insert(&bookmark); err != nil {
		t.Error(err)
	}

	session := storage.session.Copy()
	defer session.Close()
	col := session.DB("").C("bookmarks")
	var result model.Bookmark
	if err := col.FindId(bookmark.ID).One(&result); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, bookmark) {
		t.Errorf("bookmark = %v; want: %v", bookmark, result)
	}
}

func TestDelete(t *testing.T) {
	id := bson.NewObjectId()
	bookmark := model.Bookmark{ID: id}

	session := storage.session.Copy()
	defer session.Close()
	col := session.DB("").C("bookmarks")
	if err := col.Insert(bookmark); err != nil {
		t.Error(err)
	}

	if err := storage.Delete(id.Hex()); err != nil {
		t.Error(err)
	}

	if err := col.FindId(id).One(&bookmark); err.Error() != "not found" {
		t.Errorf("FindID error = %q; want: %q", err.Error(), "not found")
	}

}

func TestUpdate(t *testing.T) {
	id := bson.NewObjectId()
	bookmark := model.Bookmark{ID: id}

	session := storage.session.Copy()
	defer session.Close()
	col := session.DB("").C("bookmarks")
	if err := col.Insert(bookmark); err != nil {
		t.Error(err)
	}

	bookmark.Title = "Updated Title"
	bookmark.URL = "http://example.com"
	bookmark.Tags = []string{"how", "now"}

	if err := storage.Update(&bookmark); err != nil {
		t.Error(err)
	}

	var result model.Bookmark
	if err := col.FindId(bookmark.ID).One(&result); err != nil {
		t.Error(err)
	}
	bookmark.Timestamp = result.Timestamp // update resets timestamp
	if !reflect.DeepEqual(result, bookmark) {
		t.Errorf("bookmark = %v; want: %v", bookmark, result)
	}
}

func loadTestData() {
	fi, err := os.Open("testdata/bookmarks.json.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()
	gz, err := gzip.NewReader(fi)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(gz)
	if err != nil {
		log.Fatal(err)
	}
	var bookmarks []model.Bookmark
	err = json.Unmarshal(b, &bookmarks)
	if err != nil {
		log.Fatal(err)
	}

	session := storage.session.Copy()
	defer session.Close()

	col := session.DB("").C("bookmarks")
	err = col.DropCollection()
	if err != nil {
		log.Println(err)
	}
	bulk := col.Bulk()
	for _, bm := range bookmarks {
		bulk.Insert(bm)
	}
	_, err = bulk.Run()
	if err != nil {
		log.Fatal(err)
	}
	var bm model.Bookmark
	err = col.Find(nil).One(&bm)
	if err != nil {
		log.Fatal(err)
	}
	oid = bm.GetID()
}
