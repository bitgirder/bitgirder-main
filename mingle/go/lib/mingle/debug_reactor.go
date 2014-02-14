package mingle

import (
    "fmt"
//    "log"
)

type DebugLogger interface {
    Log( msg string )
}

type DebugLoggerFunc func( string )

func ( f DebugLoggerFunc ) Log( msg string ) { f( msg ) }

type DebugReactor struct { 
    l DebugLogger 
    Label string
}

func ( dr *DebugReactor ) ProcessEvent( ev ReactorEvent ) error {
    msg := EventToString( ev )
    if dr.Label != "" { msg = fmt.Sprintf( "[%s] %s", dr.Label, msg ) }
    dr.l.Log( msg )
    return nil
}

func NewDebugReactor( l DebugLogger ) *DebugReactor { 
    return &DebugReactor{ l: l }
}
