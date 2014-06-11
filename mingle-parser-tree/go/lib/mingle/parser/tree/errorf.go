package tree

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/parser/tree: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/parser/tree: " + tmpl, argv... )
}
