[![Go Reference](https://pkg.go.dev/badge/github.com/ugent-library/bind.svg)](https://pkg.go.dev/github.com/ugent-library/bind)

# ugent-library/bind

Package bind contains convenience functions to decode HTTP request data.

It can bind header values, router path variables, query parameters, form data
and a json or xml body to a struct.

The package uses [go-playground/form](https://github.com/go-playground/form)
under the hood for header, form and query decoding.

## Install

```sh
go get -u github.com/ugent-library/bind
```
## Examples

```go
    type UserForm struct {
        ID        int    `path:"user_id"`
        FirstName string `form:"first_name" query:"first_name" json:"first_name"`
        LastName  string `form:"last_name" query:"last_name" json:"last_name"`
    }

    // if the struct implements bind.Validator, ValidateBind() will be called
    // after a successful bind.Request()
    func (f *UserForm) ValidateBind() error {
        if f.LastName == "" {
            return errors.New("validation failed: last name can't be empty")
        }        
    }

    http.HandleFunc("/echo/user", func(w http.ResponseWriter, r *http.Request) {
        u := UserForm{}
        if err := bind.Request(r, &u); err != nil {
            // handle error
        }
        fmt.Fprintf(w, "%d: %s %s", u.ID, u.FirstName, u.LastName)
    })
```
