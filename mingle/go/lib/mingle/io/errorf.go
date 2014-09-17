package io

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/io: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/io: " + tmpl, argv... )
}
