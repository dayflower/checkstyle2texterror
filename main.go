package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/phayes/checkstyle"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	OutputSeverity bool `short:"s" long:"severity" description:"output severity (default: false)"`
}

type errorContent struct {
	line     int
	column   int
	severity string
	message  string
}

type errorFile struct {
	contents map[string]*errorContent
}

func newErrorFile() *errorFile {
	return &errorFile{contents: map[string]*errorContent{}}
}

type errorsContainer struct {
	files map[string]*errorFile
}

func newErrorsContainer() *errorsContainer {
	return &errorsContainer{map[string]*errorFile{}}
}

func (errs *errorsContainer) printErrors(outputSeverity bool) {
	filenames := make([]string, 0, len(errs.files))
	for k := range errs.files {
		filenames = append(filenames, k)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		errors := errs.files[filename]

		keys := make([]string, 0, len(errors.contents))
		for k := range errors.contents {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			contents := errors.contents[k]

			message := strings.Replace(contents.message, "\n", "", -1)

			if outputSeverity {
				fmt.Printf("%s:%d:%d:%s: %s\n", filename, contents.line, contents.column, contents.severity, message)
			} else {
				fmt.Printf("%s:%d:%d: %s\n", filename, contents.line, contents.column, message)
			}
		}
	}
}

func severityToStr(severity checkstyle.Severity) string {
	switch severity {
	case checkstyle.SeverityError:
		return "e"
	case checkstyle.SeverityWarning:
		return "w"
	case checkstyle.SeverityInfo:
		return "i"
	case checkstyle.SeverityIgnore:
		return ""
	case checkstyle.SeverityNone:
		return ""
	default:
		panic("Unsupported severity " + severity)
	}
}

func (errs *errorsContainer) addError(filename string, line int, column int, severity checkstyle.Severity, message string) {
	severityStr := severityToStr(severity)
	if severityStr == "" {
		return
	}

	file, ok := errs.files[filename]
	if !ok {
		file = newErrorFile()
		errs.files[filename] = file
	}

	// to prevent duplication
	key := fmt.Sprintf("%08d:%08d:%m", line, column, message)
	_, ok = file.contents[key]
	if !ok {
		file.contents[key] = &errorContent{line, column, severityStr, message}
	}
}

type checkstyleErrorTranslator struct {
	errors *errorsContainer
}

func (etr *checkstyleErrorTranslator) addError(file *checkstyle.File, err *checkstyle.Error) {
	etr.errors.addError(file.Name, err.Line, err.Column, err.Severity, err.Message)
}

func (etr *checkstyleErrorTranslator) parseCheckstyleErrors(checkstyle checkstyle.CheckStyle) {
	for _, file := range checkstyle.File {
		for _, err := range file.Error {
			etr.addError(file, err)
		}
	}
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
		return
	}

	errors := newErrorsContainer()

	decoder := xml.NewDecoder(os.Stdin)
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			return
		}

		switch token := t.(type) {
		case xml.StartElement:
			if token.Name.Local == "checkstyle" {
				document := checkstyle.CheckStyle{}
				decoder.DecodeElement(&document, &token)

				translator := checkstyleErrorTranslator{errors}
				translator.parseCheckstyleErrors(document)
			}
		}
	}

	errors.printErrors(opts.OutputSeverity)
}
