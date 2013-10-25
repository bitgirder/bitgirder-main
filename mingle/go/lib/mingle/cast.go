package mingle

import (
    "strings"
    "strconv"
    "encoding/base64"
    "fmt"
    "bitgirder/objpath"
)

type ValueCastError struct {
    ve ValueErrorImpl
    msg string
}

func ( e *ValueCastError ) Message() string { return e.msg }

func ( e *ValueCastError ) Error() string { 
    return e.ve.MakeError( e.Message() ) 
}

func ( e *ValueCastError ) Location() objpath.PathNode { 
    return e.ve.Location() 
}

func NewValueCastError( path idPath, msg string ) *ValueCastError {
    res := &ValueCastError{ msg: msg, ve: ValueErrorImpl{} }
    res.ve.Path = path
    return res
}

func NewValueCastErrorf(
    path idPath, tmpl string, args ...interface{} ) *ValueCastError {
    return NewValueCastError( path, fmt.Sprintf( tmpl, args... ) )
}

func NewTypeCastError( 
    expct, act TypeReference, path objpath.PathNode ) *ValueCastError {
    return NewValueCastErrorf( 
        path,
        "Expected value of type %s but found %s",
        expct.ExternalForm(), act.ExternalForm(),
    )
}

func NewTypeCastErrorValue( 
    t TypeReference, val Value, path objpath.PathNode ) *ValueCastError {
    return NewTypeCastError( t, TypeOf( val ), path )
}

func strToBool( 
    s String, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch lc := strings.ToLower( string( s ) ); lc { 
    case "true": return Boolean( true ), nil
    case "false": return Boolean( false ), nil
    }
    errTmpl :="Invalid boolean value: %s"
    errStr := QuoteValue( s )
    return nil, NewValueCastErrorf( path, errTmpl, errStr )
}

func castBoolean( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Boolean: return v, nil
    case String: return strToBool( v, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castBuffer( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Buffer: return v, nil
    case String: 
        buf, err := base64.StdEncoding.DecodeString( string( v ) )
        if err == nil { return Buffer( buf ), nil }
        msg := "Invalid base64 string: %s"
        return nil, NewValueCastErrorf( path, msg, err.Error() )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castString( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case String: return mgVal, nil
    case Boolean, Int32, Int64, Uint32, Uint64, Float32, Float64:
        return String( v.( fmt.Stringer ).String() ), nil
    case Timestamp: return String( v.Rfc3339Nano() ), nil
    case Buffer:
        return String( base64.StdEncoding.EncodeToString( []byte( v ) ) ), nil
    case *Enum: return String( v.Value.ExternalForm() ), nil
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func valueCastErrorForNumError(
    path idPath, at *AtomicTypeReference, err *strconv.NumError ) error {
    return NewValueCastErrorf( path, "%s: %s", err.Err.Error(), err.Num )
}

func parseIntInitial(
    s String,
    bitSize int,
    numType TypeReference ) ( sInt int64, uInt uint64, err error ) {
    if indx := strings.IndexAny( string( s ), "eE." ); indx >= 0 {
        var f float64
        f, err = strconv.ParseFloat( string( s ), 64 )
        if err == nil { sInt, uInt  = int64( f ), uint64( f ) }
    } else { 
        if numType == TypeUint32 || numType == TypeUint64 {
            uInt, err = strconv.ParseUint( string( s ), 10, bitSize )
            sInt = int64( uInt ) // do this even if err != nil
        } else {
            sInt, err = strconv.ParseInt( string( s ), 10, bitSize )
            uInt = uint64( sInt )
        }
    }
    return
}

func parseInt( 
    s String, 
    bitSize int, 
    numTyp TypeReference, 
    at *AtomicTypeReference, 
    path idPath ) ( Value, error ) {
    sInt, uInt, err := parseIntInitial( s, bitSize, numTyp )
    if err == nil {
        switch numTyp {
        case TypeInt32: return Int32( sInt ), nil
        case TypeInt64: return Int64( sInt ), nil
        case TypeUint32: return Uint32( uInt ), nil
        case TypeUint64: return Uint64( uInt ), nil
        default:
            msg := "Unhandled number type: %s"
            panic( NewValueCastErrorf( path, msg, numTyp ) )
        }
    } 
    if ne, ok := err.( *strconv.NumError ); ok {
        return nil, valueCastErrorForNumError( path, at, ne )
    }
    return nil, NewValueCastErrorf( path, err.Error() )
}

func castInt32( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return v, nil
    case Int64: return Int32( v ), nil
    case Uint32: return Int32( int32( v ) ), nil
    case Uint64: return Int32( int32( v ) ), nil
    case Float32: return Int32( int32( v ) ), nil
    case Float64: return Int32( int32( v ) ), nil
    case String: return parseInt( v, 32, TypeInt32, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castInt64( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Int64( v ), nil
    case Int64: return v, nil
    case Uint32: return Int64( int64( v ) ), nil
    case Uint64: return Int64( int64( v ) ), nil
    case Float32: return Int64( int64( v ) ), nil
    case Float64: return Int64( int64( v ) ), nil
    case String: return parseInt( v, 64, TypeInt64, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castUint32(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Uint32( uint32( v ) ), nil
    case Uint32: return v, nil
    case Int64: return Uint32( uint32( v ) ), nil
    case Uint64: return Uint32( uint32( v ) ), nil
    case Float32: return Uint32( uint32( v ) ), nil
    case Float64: return Uint32( uint32( v ) ), nil
    case String: return parseInt( v, 32, TypeUint32, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castUint64(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Uint64( uint64( v ) ), nil
    case Uint32: return Uint64( uint64( v ) ), nil
    case Int64: return Uint64( uint64( v ) ), nil
    case Uint64: return v, nil
    case Float32: return Uint64( uint64( v ) ), nil
    case Float64: return Uint64( uint64( v ) ), nil
    case String: return parseInt( v, 64, TypeUint64, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func parseFloat32(
    s string,
    bitSize int,
    numTyp TypeReference,
    at *AtomicTypeReference,
    path idPath ) ( Value, error ) {
    f, err := strconv.ParseFloat( string( s ), bitSize )
    if err != nil { 
        ne := err.( *strconv.NumError )
        return nil, valueCastErrorForNumError( path, at, ne )
    }
    switch numTyp {
    case TypeFloat32: return Float32( f ), nil
    case TypeFloat64: return Float64( f ), nil
    }
    panic( NewValueCastErrorf( path, "Unhandled num type: %s", numTyp ) )
}

func castFloat32( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Float32( float32( v ) ), nil
    case Int64: return Float32( float32( v ) ), nil
    case Uint32: return Float32( float32( v ) ), nil
    case Uint64: return Float32( float32( v ) ), nil
    case Float32: return v, nil
    case Float64: return Float32( float32( v ) ), nil
    case String: return parseFloat32( string( v ), 32, TypeFloat32, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castFloat64( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Int32: return Float64( float64( v ) ), nil
    case Int64: return Float64( float64( v ) ), nil
    case Uint32: return Float64( float64( v ) ), nil
    case Uint64: return Float64( float64( v ) ), nil
    case Float32: return Float64( float64( v ) ), nil
    case Float64: return v, nil
    case String: return parseFloat32( string( v ), 64, TypeFloat64, at, path )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castTimestamp( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case Timestamp: return v, nil
    case String:
        tm, err := ParseTimestamp( string( v ) )
        if err == nil { return tm, nil }
        msg := "Invalid timestamp: %s"
        return nil, NewValueCastErrorf( path, msg, err.Error() )
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castEnum( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case *Enum: if v.Type.Equals( at.Name ) { return v, nil }
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castSymbolMap( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    switch v := mgVal.( type ) {
    case *SymbolMap: return v, nil
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func castNull( 
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    if _, ok := mgVal.( *Null ); ok { return mgVal, nil }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

// switch compares based on qname not at itself since we may be dealing with
// restriction types, meaning that if at is mingle:core@v1/String~"a", it is a
// string (has qname mingle:core@v1/String) but will not equal TypeString itself
func castAtomicUnrestricted(
    mgVal Value, at *AtomicTypeReference, path idPath ) ( Value, error ) {
    if _, ok := mgVal.( *Null ); ok {
        if at.Equals( TypeNull ) { return mgVal, nil }
        return nil, NewValueCastErrorf( path, "Value is null" )
    }
    switch nm := at.Name; {
    case nm.Equals( QnameBoolean ): return castBoolean( mgVal, at, path )
    case nm.Equals( QnameBuffer ): return castBuffer( mgVal, at, path )
    case nm.Equals( QnameString ): return castString( mgVal, at, path )
    case nm.Equals( QnameInt32 ): return castInt32( mgVal, at, path )
    case nm.Equals( QnameInt64 ): return castInt64( mgVal, at, path )
    case nm.Equals( QnameUint32 ): return castUint32( mgVal, at, path )
    case nm.Equals( QnameUint64 ): return castUint64( mgVal, at, path )
    case nm.Equals( QnameFloat32 ): return castFloat32( mgVal, at, path )
    case nm.Equals( QnameFloat64 ): return castFloat64( mgVal, at, path )
    case nm.Equals( QnameTimestamp ): return castTimestamp( mgVal, at, path )
    case nm.Equals( QnameSymbolMap ): return castSymbolMap( mgVal, at, path )
    case nm.Equals( QnameNull ): return castNull( mgVal, at, path )
    case nm.Equals( QnameValue ): return mgVal, nil
    }
    if _, ok := mgVal.( *Enum ); ok { return castEnum( mgVal, at, path ) }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func checkRestriction( val Value, at *AtomicTypeReference, path idPath ) error {
    if at.Restriction.AcceptsValue( val ) { return nil }
    return NewValueCastErrorf( 
        path, "Value %s does not satisfy restriction %s",
        QuoteValue( val ), at.Restriction.ExternalForm() )
}

func castAtomic(
    mgVal Value, 
    at *AtomicTypeReference,
    path idPath ) ( val Value, err error ) {
    if val, err = castAtomicUnrestricted( mgVal, at, path ); err == nil {
        if at.Restriction != nil { err = checkRestriction( val, at, path ) }
    }
    return
}
