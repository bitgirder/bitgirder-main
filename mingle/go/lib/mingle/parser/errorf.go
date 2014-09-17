package parser

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/parser: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/parser: " + tmpl, argv... )
}
