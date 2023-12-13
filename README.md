# ❤️ Beautiful Web Error in GoLang
Simple output of information about an error request and output of a convenient traceback view

## Preview
![image](https://github.com/xALEGORx/beautiful-web-error/assets/26199734/df0f6382-c9c3-4341-80c2-95b705ce2f53)

## How to install?
```bash
# Install package
go get -u github.com/xALEGORx/beautiful-web-error
```

## Using
`In your main function, declare the debugger structure and run the initialization function`
```go
import (
  bwError "github.com/xALEGORx/beautiful-web-error"
)

var berror bwError.BeautifulError = bwError.BeautifulError{
	Page:          true,            // required, is the HTML page display enabled
	Theme:         "xcode-dark",    // theme for highlighter code, default xcode-dark
	IgnoredErrors: []error{io.EOF}, // ignoring errors for debugging, default empty
}

func main() {
  // Init BeautifulError
  if err := berror.Init(); err != nil {
    panic(err)
  }
}
```
`Next, in the right place to debug the error, call the Catch function`
```go
func handler(w http.ResponseWriter, r *http.Request) {
  if err := makeRequest(r); berror.Catch(err, w, r) {
    return
  }
}
```

## Disable HTML
When releasing to production, so that the HTML debugging page is not displayed. You can set the Page: false parameter. 
Then the error output will be as follows:
```
{"error":"failed request"}
```
