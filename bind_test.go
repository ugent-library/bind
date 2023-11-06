package bind

import (
	"net/http"
	"testing"
)

func TestPath(t *testing.T) {
	PathValueFunc = func(r *http.Request, k string) string {
		if k == "id" {
			return "123"
		}
		return ""
	}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	v1 := struct {
		ID string `path:"id"`
	}{}
	if err := Path(r, &v1); err != nil {
		t.Error(err)
	}
	if v1.ID != "123" {
		t.Errorf("got %q, want %q", v1.ID, "123")
	}
	if err := Path(r, v1); err != ErrInvalidType {
		t.Error("Path should return ErrInvalidType unless target is a non-nil pointer to struct")
	}
}
