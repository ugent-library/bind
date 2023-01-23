[![Go Reference](https://pkg.go.dev/badge/github.com/ugent-library/bind.svg)](https://pkg.go.dev/github.com/ugent-library/bind)

# ugent-library/bind

Package bind contains convenience functions to decode HTTP request data.

It uses [go-playground/form](https://github.com/go-playground/form) under the hood.

## Install

```sh
go get -u ugent-library/bind
```
## Examples

```go
    type UserForm struct {
        FirstName `form:"first_name" query:"first_name"`
        LastName `form:"last_name" query:"last_name"`
    }

    http.HandleFunc("/echo_name", func(w http.ResponseWriter, r *http.Request) {
        u := UserForm{}
        if err := bind.Request(r, &u); err != nil {
            // handle error
        }
        fmt.Fprintf(w, "%s %s", u.FirstName, u.LastName)
    })
```
