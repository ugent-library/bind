package bind

import (
	"net/http"
	"net/url"
	"testing"
)

func TestPath(t *testing.T) {
	PathFunc = func(r *http.Request) url.Values {
		return url.Values{"id": []string{"123"}}
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
	if err := Path(r, v1); err == nil {
		t.Error("Path should return an error unless target is a non-nil pointer")
	}
}
