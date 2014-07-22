package parser

import (
    "fmt"
)

const ParseSourceInput = "<input>"

type Location struct {
    Line, Col int
    Source string
}

func ( l *Location ) Dup() *Location {
    return &Location{ Line: l.Line, Col: l.Col, Source: l.Source }
}

func ( l *Location ) String() string {
    return fmt.Sprintf( "[%s, line %d, col %d]", l.Source, l.Line, l.Col )
}

type ParseError struct {
    Message string
    Loc *Location
}

func ( e *ParseError ) Error() string {
    return fmt.Sprintf( "%s: %s", e.Loc, e.Message )
}

func NewParseErrorf( 
    loc *Location, tmpl string, argv ...interface{} ) *ParseError {

    return &ParseError{ Loc: loc, Message: fmt.Sprintf( tmpl, argv... ) }
}
