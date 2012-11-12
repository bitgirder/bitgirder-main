package bincodec

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "mingle/codec/bincodec: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "mingle/codec/bincodec: " + tmpl, argv... )
}
