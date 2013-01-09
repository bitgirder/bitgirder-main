package json

import (
    "fmt"
    mg "mingle"
    "mingle/codec"
    "mingle/parser/loc"
    "io"
    "strings"
    "container/list"
//    "log"
    gojson "encoding/json"
    "encoding/base64"
    "bitgirder/objpath"
)

const (
    jsonKeyType = "$type"
    jsonKeyConstant = "$constant"
)

var CodecId = mg.MustIdentifier( "json" )

func newCodecErrorf( 
    loc objpath.PathNode, msg string, args ...interface{} ) error {
    argsAct := args
    if loc != nil {
        msg = "%s: " + msg
        argsAct = make( []interface{}, 1, len( args ) + 1 )
        argsAct[ 0 ] = objpath.Format( loc, objpath.StringDotFormatter )
        argsAct = append( argsAct, args... )
    }
    return codec.Errorf( msg, argsAct... )
}

func descendInbound( p objpath.PathNode, k string ) objpath.PathNode {
    if p == nil {
        return objpath.RootedAt( k )
    }
    return p.Descend( k )
}

func startList( p objpath.PathNode ) *objpath.ListNode {
    if p == nil { return objpath.RootedAtList() }
    return p.StartList()
}

func parseErrorMessageOf( err error ) string {
    return err.( *loc.ParseError ).Message
}

const (
    tmplInvalidFieldId = "Invalid field name %q: %s"
    tmplInvalidEnumVal = "Invalid enum value %q: %s"
)

func expectIdentifier( 
    s, tmpl string, errLoc objpath.PathNode ) ( *mg.Identifier, error ) {
    id, err := mg.ParseIdentifier( s )
    if err == nil { return id, nil }
    msg := parseErrorMessageOf( err )
    return nil, newCodecErrorf( errLoc, tmpl, s, msg )
}

type JsonCodecOpts struct {
    IdFormat mg.IdentifierFormat
    ExpandEnums bool
    OmitTypeFields bool
}

var defaultCodecOpts = &JsonCodecOpts{
    IdFormat: mg.LcHyphenated,
}

type JsonCodec struct {
    opts JsonCodecOpts
}

func ( c *JsonCodec ) asJsonEnum( en *mg.Enum ) interface{} {
    valStr := en.Value.ExternalForm()
    if c.opts.ExpandEnums {
        m := make( map[ string ]interface{}, 2 )
        m[ jsonKeyType ] = en.Type.ExternalForm()
        m[ jsonKeyConstant ] = valStr
        return m
    }
    return valStr
}

func ( c *JsonCodec ) asJsonValue( val mg.Value ) interface{} {
    switch v := val.( type ) {
    case mg.String: return string( v )
    case mg.Boolean: return bool( v )
    case mg.Int32: return int32( v )
    case mg.Int64: return int64( v )
    case mg.Uint32: return uint32( v )
    case mg.Uint64: return uint64( v )
    case mg.Float32: return float32( v )
    case mg.Float64: return float64( v )
    case mg.Buffer: return base64.StdEncoding.EncodeToString( v )
    case *mg.Enum: return c.asJsonEnum( v )
    case mg.Timestamp: return v.Rfc3339Nano()
    case *mg.Null: return nil
    }
    panic( libErrorf( "Unhandled mingle value: %T", val ) )
}

type encoder struct {
    c *JsonCodec
    w io.Writer
    stack *list.List
}

// Both value() and end() operate on and return go json vals (not mg.Value)
type accumulator interface {
    value( v interface{} )
    end() interface{}
}

type mapAcc struct {
    m map[ string ]interface{}
    fld *mg.Identifier
    f mg.IdentifierFormat
}

func ( e *encoder ) newMapAcc() *mapAcc { 
    return &mapAcc{ f: e.c.opts.IdFormat, m: make( map[ string ]interface{} ) }
}

func ( m *mapAcc ) value( v interface{} ) {
    m.m[ m.fld.Format( m.f ) ] = v
    m.fld = nil
}

func ( m *mapAcc ) end() interface{} { return m.m }

type listAcc struct { l []interface{} }

func newListAcc() *listAcc { return &listAcc{ make( []interface{}, 0, 4 ) } }

func ( l *listAcc ) value( v interface{} ) { l.l = append( l.l, v ) }

func ( l *listAcc ) end() interface{} { return l.l }

func ( e *encoder ) push( acc accumulator ) { e.stack.PushFront( acc ) }

func ( e *encoder ) peek() accumulator { 
    return e.stack.Front().Value.( accumulator ) 
}

func ( e *encoder ) pop() accumulator {
    res := e.peek()
    e.stack.Remove( e.stack.Front() )
    return res
}

func ( e *encoder ) startStruct( typ mg.TypeReference ) {
    acc := e.newMapAcc()
    if ! e.c.opts.OmitTypeFields { acc.m[ jsonKeyType ] = typ.ExternalForm() }
    e.push( acc )
}

func ( e *encoder ) startList() { e.push( newListAcc() ) }

func ( e *encoder ) startMap() { e.push( e.newMapAcc() ) }

func ( e *encoder ) startField( fld *mg.Identifier ) {
    e.peek().( *mapAcc ).fld = fld
}

func ( e *encoder ) value( val mg.Value ) {
    e.peek().value( e.c.asJsonValue( val ) )
}

func ( e *encoder ) end() error {
    val := e.pop().end()
    if e.stack.Len() == 0 {
        enc := gojson.NewEncoder( e.w )
        return enc.Encode( val )
    }
    e.peek().value( val )
    return nil
}

func ( e *encoder ) ProcessEvent( ev mg.ReactorEvent ) error {
    switch v := ev.( type ) {
    case mg.ValueEvent: e.value( v.Val )
    case mg.ListStartEvent: e.startList()
    case mg.StructStartEvent: e.startStruct( v.Type )
    case mg.MapStartEvent: e.startMap()
    case mg.FieldStartEvent: e.startField( v.Field )
    case mg.EndEvent: return e.end()
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return nil
}

func ( c *JsonCodec ) EncoderTo( w io.Writer ) mg.ReactorEventProcessor {
    return &encoder{ 
        w: w, 
        c: c, 
        stack: &list.List{},
    }
}

var numCastRootPath objpath.PathNode
func init() { numCastRootPath = objpath.RootedAt( "number" ) }

func asMingleNumber( n gojson.Number ) ( mg.Value, error ) {
    var typ mg.TypeReference
    if i := strings.IndexAny( string( n ), ".eE" ); i >= 0 {
        typ = mg.TypeFloat64
    } else { typ = mg.TypeInt64 }
    return mg.CastValue( mg.String( string( n ) ), typ, numCastRootPath )
}

func visitError( path objpath.PathNode, msg string ) error {
    if path != nil {
        msg = objpath.Format( path, objpath.StringDotFormatter ) + ": " + msg
    }
    return codec.Error( msg )
}

func visitErrorf( 
    path objpath.PathNode, tmpl string, argv ...interface{} ) error {
    return visitError( path, fmt.Sprintf( tmpl, argv... ) )
}

func ( c *JsonCodec ) visitNumber(
    n gojson.Number,
    path objpath.PathNode,
    rep mg.ReactorEventProcessor ) error {
    mgNum, err := asMingleNumber( n )
    if err != nil { return err }
    return rep.ProcessEvent( mg.ValueEvent{ mgNum } )
}

func ( c *JsonCodec ) visitValue(
    goVal interface{}, 
    path objpath.PathNode, 
    rep mg.ReactorEventProcessor ) error {
    switch v := goVal.( type ) {
    case nil: return rep.ProcessEvent( mg.ValueEvent{ mg.NullVal } )
    case gojson.Number: return c.visitNumber( v, path, rep )
    case string: return rep.ProcessEvent( mg.ValueEvent{ mg.String( v ) } )
    case bool: return rep.ProcessEvent( mg.ValueEvent{ mg.Boolean( v ) } )
    case map[ string ]interface{}: return c.visitMap( v, path, rep )
    case []interface{}: return c.visitList( v, path, rep )
    }
    panic( libErrorf( "Unhandled json go value: %T", goVal ) )
}

func ( c *JsonCodec ) visitList(
    l []interface{}, 
    path objpath.PathNode, 
    rep mg.ReactorEventProcessor ) error {
    lp := startList( path )
    if err := rep.ProcessEvent( mg.EvListStart ); err != nil { return err }
    for _, val := range l {
        if err := c.visitValue( val, lp, rep ); err != nil { return err }
        lp = lp.Next()
    }
    return rep.ProcessEvent( mg.EvEnd )
}

// m[ $constant ] known to be present, though not necessarily valid
func ( c *JsonCodec ) visitEnum(
    typ mg.TypeReference,
    m map[ string ]interface{},
    path objpath.PathNode,
    rep mg.ReactorEventProcessor ) error {
    if len( m ) > 2 {
        return visitError( path, "Enum has one or more unrecognized keys" )
    }
    var val *mg.Identifier
    if valStr, ok := m[ jsonKeyConstant ].( string ); ok {
        errLoc := descendInbound( path, jsonKeyConstant )
        var err error
        val, err = expectIdentifier( valStr, tmplInvalidEnumVal, errLoc )
        if err != nil { return err }
    } else { 
        errLoc := descendInbound( path, jsonKeyConstant )
        return visitError( errLoc, "Invalid constant value" )
    }
    return rep.ProcessEvent( mg.ValueEvent{ &mg.Enum{ typ, val } } )
}

func ( c *JsonCodec ) visitFields(
    m map[ string ]interface{},
    path objpath.PathNode,
    rep mg.ReactorEventProcessor ) error {
    for fld, val := range m {
        if fld != jsonKeyType {
            if len( fld ) > 0 && fld[ 0 ] == byte( '$' ) {
                return visitErrorf( path, "Unrecognized control key: %q", fld )
            }
            fldPath := descendInbound( path, fld )
            id, err := expectIdentifier( fld, tmplInvalidFieldId, fldPath )
            if err != nil { return err }
            if val != nil {
                ev := mg.FieldStartEvent{ id }
                if err = rep.ProcessEvent( ev ); err != nil { return err }
                valPath := descendInbound( path, fld )
                if err = c.visitValue( val, valPath, rep ); err != nil { 
                    return err
                }
            }
        }
    }
    return rep.ProcessEvent( mg.EvEnd )
}

func parseAtomicType( 
    typStr string, 
    path objpath.PathNode, 
    key string ) ( *mg.AtomicTypeReference, error ) {
    typ, err := mg.ParseTypeReference( typStr )
    if err == nil {
        if at, atOk := typ.( *mg.AtomicTypeReference ); atOk { return at, nil }
        errPath := descendInbound( path, key )
        return nil, visitError( errPath, "not an atomic type" )
    }
    errStr := parseErrorMessageOf( err )
    return nil, visitError( descendInbound( path, key ), errStr )
}

func ( c *JsonCodec ) visitMap( 
    m map[ string ]interface{}, 
    path objpath.PathNode, 
    rep mg.ReactorEventProcessor ) error {
    if typVal, ok := m[ jsonKeyType ]; ok {
        if typStr, ok2 := typVal.( string ); ok2 {
            at, err := parseAtomicType( typStr, path, jsonKeyType )
            if err != nil { return err }
            if _, ok4 := m[ jsonKeyConstant ]; ok4 {
                return c.visitEnum( at, m, path, rep )
            }
            ev := mg.StructStartEvent{ at }
            if err = rep.ProcessEvent( ev ); err != nil { return err }
        }
    } else {
        if path == nil {
            return visitErrorf( path, "Missing type key (%q)", jsonKeyType )
        }
        if err := rep.ProcessEvent( mg.EvMapStart ); err != nil { return err }
    }
    return c.visitFields( m, path, rep )
}

func ( c *JsonCodec ) DecodeFrom( 
    r io.Reader, rep mg.ReactorEventProcessor ) error {
    dec := gojson.NewDecoder( r )
    dec.UseNumber()
    var dest interface{}
    if err := dec.Decode( &dest ); err != nil { return err }
    if m, ok := dest.( map[ string ]interface{} ); ok {
        return c.visitMap( m, nil, rep )
    }
    return codec.Errorf( "Unexpected top level JSON value" )
}

type JsonCodecInitializerError struct {
    msg string
}

func ( e *JsonCodecInitializerError ) Error() string { return e.msg }

func CreateJsonCodec( opts *JsonCodecOpts ) ( *JsonCodec, error ) {
    res := &JsonCodec{ *opts } // res now has copy of opts
    if opts.IdFormat == 0 { res.opts.IdFormat = mg.LcHyphenated }
    if opts.ExpandEnums && opts.OmitTypeFields {
        msg := "Invalid combination: ExpandEnums and OmitTypeFields"
        return nil, &JsonCodecInitializerError{ msg }
    }
    return res, nil
}

func MustJsonCodec( opts *JsonCodecOpts ) *JsonCodec {
    res, err := CreateJsonCodec( opts )
    if err != nil { panic( err ) }
    return res
}

func NewJsonCodec() *JsonCodec { return MustJsonCodec( defaultCodecOpts ) }

func init() {
    codec.RegisterCodec(
        &codec.CodecRegistration{
            Codec: NewJsonCodec(),
            Id: CodecId,
            Source: "mingle/codec/json",
        },
    )
}
