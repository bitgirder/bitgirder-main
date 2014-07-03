package mingle

import (
    "bytes"
    "fmt"
)

type valueQuote struct {
    buf *bytes.Buffer
}

func ( vq valueQuote ) appendEnum( e *Enum ) {
    fmt.Fprintf( vq.buf, "%s.%s", 
        e.Type.ExternalForm(), e.Value.ExternalForm() )
}

func ( vq valueQuote ) appendList( l *List ) {
    vq.buf.WriteRune( '[' )
    for i, val := range l.vals {
        vq.appendValue( val )
        if i < len( l.vals ) - 1 { vq.buf.WriteString( ", " ) }
    }
    vq.buf.WriteRune( ']' )
}

func ( vq valueQuote ) appendSymbolMapFields( m *SymbolMap ) {
    vq.buf.WriteRune( '{' )
    remain := m.Len() - 1
    m.EachPair( func( fld *Identifier, val Value ) {
        vq.buf.WriteString( fld.Format( LcCamelCapped ) )
        vq.buf.WriteRune( ':' )
        vq.appendValue( val )
        if remain > 0 { vq.buf.WriteString( ", " ) }
        remain--
    })
    vq.buf.WriteRune( '}' )
}

func ( vq valueQuote ) appendSymbolMap( m *SymbolMap ) {
    vq.appendSymbolMapFields( m )
}

func ( vq valueQuote ) appendStruct( ms *Struct ) {
    vq.buf.WriteString( ms.Type.ExternalForm() )
    vq.appendSymbolMapFields( ms.Fields )
}

func ( vq valueQuote ) appendValue( val Value ) {
    switch v := val.( type ) {
    case String: fmt.Fprintf( vq.buf, "%q", string( v ) )
    case Buffer: fmt.Fprintf( vq.buf, "buf[%d]", len( []byte( v ) ) )
    case Timestamp: fmt.Fprintf( vq.buf, "%s", v.Rfc3339Nano() )
    case *Null: vq.buf.WriteString( "null" )
    case Boolean, Int32, Int64, Uint32, Uint64, Float32, Float64:
        vq.buf.WriteString( val.( fmt.Stringer ).String() )
    case *Enum: vq.appendEnum( v )
    case *List: vq.appendList( v )
    case *SymbolMap: vq.appendSymbolMap( v )
    case *Struct: vq.appendStruct( v )
    default: fmt.Fprintf( vq.buf, "(!%T)", val ) // seems better than a panic
    }
}

func QuoteValue( val Value ) string { 
    vq := valueQuote{ buf: &bytes.Buffer{} }
    vq.appendValue( val )
    return vq.buf.String()
}
