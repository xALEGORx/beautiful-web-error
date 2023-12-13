package beautiful_web_error

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// the structure for initializing the library
type BeautifulError struct {
	Page          bool    // is the HTML page display enabled
	Theme         string  // theme for highlighter code
	IgnoredErrors []error // ignoring errors for debugging
}

// initialization method for parsing the template and loading the theme
func (b *BeautifulError) Init() error {
	var err error

	errorTemplate, err = template.ParseFS(content, "templates/index.html")
	if err != nil {
		return err
	}

	if b.Theme == "" {
		b.Theme = "xcode-dark" // setting default theme
	}

	return nil
}

// the main method of the library that handles errors.
// If the transmitted error is not in the IgnoredErrors array, the render method will be called
func (b BeautifulError) Catch(err error, w http.ResponseWriter, r *http.Request) bool {
	if err == nil {
		return false
	}

	for _, ignoredError := range b.IgnoredErrors {
		if ignoredError == err {
			return false
		}
	}

	b.render(err, w, r)

	return true
}

///////////////////////////////////////////////////

// uploading a template via the embed library
//
//go:embed templates/index.html
var content embed.FS

// an instance of the template to be used for HTML rendering
var errorTemplate *template.Template

// a structure for displaying information about a specific file on the stack
type stackFrame struct {
	FullPath   string
	FileName   string
	LineNumber int
	Function   string
	Code       template.HTML
}

// a structure for displaying information about a request from http.Request
type requestData struct {
	Method       string
	Host         string
	URL          string
	Proto        string
	RemoteAddr   string
	UserAgent    string
	FormData     map[string]string
	PostFormData map[string]string
	Headers      map[string]string
}

// structure for error output if Page (HTML render) is disabled
type errorResponse struct {
	Error string `json:"error"`
}

// the main method for rendering a page with debugging
func (b BeautifulError) render(err error, w http.ResponseWriter, r *http.Request) {
	if !b.Page {
		// if disabled page render, render json message
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(errorResponse{
			Error: err.Error(),
		})

		return
	}

	stackTrace := b.getStackTrace()
	formatedRequestData := b.formatRequestData(r)

	data := struct {
		ErrorMessage string
		StackTrace   []stackFrame
		RequestData  requestData
	}{
		ErrorMessage: err.Error(),
		StackTrace:   stackTrace,
		RequestData:  formatedRequestData,
	}

	err = errorTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getting complete information about all files in the stack
func (b BeautifulError) getStackTrace() []stackFrame {
	pc := make([]uintptr, 20)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[2:n]) // we use shift 2: to disable the output of the library's working methods

	var stackTrace []stackFrame
	for i := 0; i < 10; i++ {
		frame, more := frames.Next()
		code := b.readCodeLines(frame.File, frame.Line)
		fileName := strings.Split(frame.File, "/")

		stackTrace = append(stackTrace, stackFrame{
			FullPath:   frame.File,
			FileName:   fileName[len(fileName)-1],
			LineNumber: frame.Line,
			Function:   frame.Function,
			Code:       template.HTML(code),
		})

		if !more {
			break
		}
	}

	return stackTrace
}

// getting a code snippet from a stack file and highlighting a line
func (b BeautifulError) readCodeLines(filename string, errorLine int) string {
	file, err := os.Open(filename)
	if err != nil {
		return "failed open file"
	}
	defer file.Close()

	startLine := errorLine - 9
	endLine := errorLine + 9

	scanner := bufio.NewScanner(file)
	lines := []string{}
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		if lineNumber >= startLine && lineNumber <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNumber > endLine {
			break
		}
	}

	code, err := b.highlightCode(strings.Join(lines, "\n"), errorLine, startLine)
	if err != nil {
		return "failed highlight code"
	}

	return code
}

// preparing query data for struct
func (b BeautifulError) formatRequestData(r *http.Request) requestData {
	r.ParseForm()

	toMapString := func(a map[string][]string) map[string]string {
		result := map[string]string{}
		for k, v := range a {
			result[k] = strings.Join(v, " ")
		}
		return result
	}

	return requestData{
		Method:       r.Method,
		Host:         r.Host,
		URL:          r.URL.Path,
		Proto:        r.Proto,
		RemoteAddr:   r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		FormData:     toMapString(r.Form),
		PostFormData: toMapString(r.PostForm),
		Headers:      toMapString(r.Header),
	}
}

// function to highlight code (formatting string code into html code)
func (b BeautifulError) highlightCode(code string, errorLine int, startLine int) (string, error) {
	var output bytes.Buffer

	lexer := lexers.Get("go")
	lexer = chroma.Coalesce(lexer)

	style := styles.Get(b.Theme)
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(
		html.WithLineNumbers(true),
		html.BaseLineNumber(startLine),
		html.HighlightLines([][2]int{{errorLine, errorLine}}),
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	if err := formatter.Format(&output, style, iterator); err != nil {
		return "", err
	}

	return output.String(), nil
}
