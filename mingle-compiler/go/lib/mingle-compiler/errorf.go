package mingle-compiler

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle-compiler: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle-compiler: " + tmpl, argv... )
}
