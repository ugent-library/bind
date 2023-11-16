package bind

import (
	"net/http"
	"testing"
)

func TestPath(t *testing.T) {
	type t1 struct {
		ID string `path:"id"`
	}

	type t2 struct {
		*t1
	}

	type t3 struct {
		t1
	}

	type t4 struct {
		ID *string `path:"id"`
	}

	type t5 struct {
		ID t1 `path:"id"`
	}

	PathValueFunc = func(r *http.Request, k string) string {
		if k == "id" {
			return "123"
		}
		return ""
	}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	v1 := t1{}
	// bind string
	if err := Path(r, &v1); err != nil {
		t.Error(err)
	} else if v1.ID != "123" {
		t.Errorf("got %q, want %q", v1.ID, "123")
	}

	// Path only accepts non nil struct pointers
	if err := Path(r, ""); err == nil {
		t.Error("got nil, want error")
	}
	if err := Path(r, v1); err == nil {
		t.Error("got nil, want error")
	}
	var ptr *t1
	if err := Path(r, ptr); err == nil {
		t.Error("got nil, want error")
	}

	v2 := t2{}
	// skip embedded struct nil pointer
	if err := Path(r, &v2); err != nil {
		t.Error(err)
	}
	// bind embedded struct pointer
	v2.t1 = &t1{}
	if err := Path(r, &v2); err != nil {
		t.Error(err)
	} else if v2.ID != "123" {
		t.Errorf("got %q, want %q", v2.ID, "123")
	}

	v3 := t3{}
	// bind embedded struct
	if err := Path(r, &v3); err != nil {
		t.Error(err)
	} else if v3.ID != "123" {
		t.Errorf("got %q, want %q", v3.ID, "123")
	}

	// bind string pointer
	str := ""
	v4 := t4{ID: &str}
	if err := Path(r, &v4); err != nil {
		t.Error(err)
	} else if *v4.ID != "123" {
		t.Errorf("got %q, want %q", *v4.ID, "123")
	}
	// bind string nil pointer
	v4.ID = nil
	if err := Path(r, &v4); err != nil {
		t.Error(err)
	} else if v4.ID == nil {
		t.Errorf("got nil, want %q", "123")

	} else if *v4.ID != "123" {
		t.Errorf("got %q, want %q", *v4.ID, "123")
	}

	// path can only bind scalar values or pointers to scalar values
	v5 := t5{}
	if err := Path(r, &v5); err == nil {
		t.Error("got nil, want error")
	}
}
