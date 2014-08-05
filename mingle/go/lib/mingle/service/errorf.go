package service

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/service: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/service: " + tmpl, argv... )
}
