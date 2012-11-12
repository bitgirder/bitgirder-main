package io

import (
    "encoding/binary"
    "bitgirder/objpath"
    "fmt"
//    "log"
    "io"
    mg "mingle"
)

const (
    HeadersVersion1 = int32( 1 )
    TypeCodeHeadersField = int32( 1 )
    TypeCodeHeadersEnd = int32( 2 )
)

var ByteOrder = binary.LittleEndian

type Headers struct {
    flds *mg.SymbolMap
}

func ( h *Headers ) Fields() *mg.SymbolMap { return h.flds }

var hdrsPath = objpath.RootedAt( mg.MustIdentifier( "headers" ) )

func ( h *Headers ) FieldsAccessor() *mg.SymbolMapAccessor {
    return mg.NewSymbolMapAccessor( h.Fields(), hdrsPath )
}

func MustHeadersPairs( pairs ...interface{} ) *Headers {
    map1 := mg.MustSymbolMap( pairs... )
    pairs2 := make( []interface{}, 0, len( pairs ) )
    map1.EachPair( func( fld *mg.Identifier, val mg.Value ) {
        path := objpath.RootedAt( fld )
        if val2, err := mg.CastValue( val, mg.TypeString, path ); err == nil {
            pairs2 = append( pairs2, fld, val2 )
        } else { panic( err ) }
    })
    return &Headers{ mg.MustSymbolMap( pairs2... ) }
}

func MustHeaders( sm *mg.SymbolMap ) *Headers {
    sm.EachPair( func( fld *mg.Identifier, val mg.Value ) {
        if _, ok := val.( mg.String ); ! ok {
            errTmpl := "Non-String value for header '%s': %T"
            panic( fmt.Errorf( errTmpl, fld, val ) )
        }
    })
    return &Headers{ sm }
}

func WriteBinary( data interface{}, w io.Writer ) error {
    return binary.Write( w, ByteOrder, data )
}

// These are preferred over bare calls to WriteBinary(), since these help detect
// errors such as passing in things of type int where int32 or int64 might be
// intended (example: len( []byte ) when writing a buffer's length ahead of its
// data).
func WriteInt32( i int32, w io.Writer ) error { return WriteBinary( i, w ) }
func WriteInt64( i int64, w io.Writer ) error { return WriteBinary( i, w ) }

func writeUtf8( s string, w io.Writer ) error {
    buf := []byte( s )
    if err := WriteInt32( int32( len( buf ) ), w ); err != nil { return err }
    if _, err := w.Write( buf ); err != nil { return err }
    return nil
}

func writeHeaderPair( k, v string, w io.Writer ) error {
    if err := WriteInt32( TypeCodeHeadersField, w ); err != nil { 
        return err 
    }
    if err := writeUtf8( k, w ); err != nil { return err }
    if err := writeUtf8( v, w ); err != nil { return err }
    return nil
}

func WriteHeaders( h *Headers, w io.Writer ) error {
    if err := WriteBinary( HeadersVersion1, w ); err != nil { return err }
    h.flds.EachPairError( func( k *mg.Identifier, v mg.Value ) error {
        return writeHeaderPair( k.ExternalForm(), string( v.( mg.String ) ), w )
    })
    return WriteInt32( TypeCodeHeadersEnd, w )
}

func readBinary( data interface{}, r io.Reader ) error {
    return binary.Read( r, ByteOrder, data )
}

func ReadInt32( r io.Reader ) ( i int32, err error ) {
    err = readBinary( &i, r )
    return
}

func ReadInt64( r io.Reader ) ( i int64, err error ) {
    err = readBinary( &i, r )
    return
}

func readUtf8( r io.Reader ) ( string, error ) {
    sz, err := ReadInt32( r )
    if err != nil { return "", err }
    buf := make( []byte, sz )
    _, err = io.ReadFull( r, buf )
    if err == nil { return string( buf ), nil }
    return "", err
}

type InvalidVersionError struct { msg string }
func ( e *InvalidVersionError ) Error() string { return e.msg }

type InvalidTypeCodeError struct { msg string }
func ( e *InvalidTypeCodeError ) Error() string { return e.msg }

func NewInvalidTypeCodeError( code int32 ) *InvalidTypeCodeError {
    msg := fmt.Sprintf( "Invalid type code: 0x%08x", code )
    return &InvalidTypeCodeError{ msg }
}

func ReadVersion( verExpct int32, verType string, r io.Reader ) error {
    ver, err := ReadInt32( r )
    if err != nil { return err }
    if ver != verExpct { 
        tmpl := "Invalid %s version: 0x%08x (expected 0x%08x)"
        msg := fmt.Sprintf( tmpl, verType, ver, verExpct )
        return &InvalidVersionError{ msg }
    }
    return nil
}

func ReadTypeCode( r io.Reader ) ( int32, error ) { return ReadInt32( r ) }

func ExpectTypeCode( expct int32, r io.Reader ) error {
    if code, err := ReadTypeCode( r ); err == nil {
        if code != expct {
            tmpl := "Invalid type code: 0x%08x (expected 0x%08x)"
            return &InvalidTypeCodeError{ fmt.Sprintf( tmpl, code, expct ) }
        }
    } else { return err }
    return nil
}

func readHeadersField( 
    pairs []interface{}, r io.Reader ) ( []interface{}, error ) {
    var keyStr, valStr string
    var err error
    if keyStr, err = readUtf8( r ); err != nil { return nil, err }
    if valStr, err = readUtf8( r ); err != nil { return nil, err }
    return append( pairs, keyStr, valStr ), nil
}

func ReadHeaders( rd io.Reader ) ( *Headers, error ) {
    if err := ReadVersion( HeadersVersion1, "headers", rd ); err != nil { 
        return nil, err 
    }
    flds := make( []interface{}, 0, 8 ) // 4 pairs typical (informal heuristic)
    err := error( nil )
    for err == nil {
        var code int32
        if code, err = ReadTypeCode( rd ); err != nil { return nil, err }
        switch code {
        case TypeCodeHeadersField: flds, err = readHeadersField( flds, rd )
        case TypeCodeHeadersEnd: return MustHeadersPairs( flds... ), nil
        default: return nil, NewInvalidTypeCodeError( code )
        }
    }
    return nil, err
}
