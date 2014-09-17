package types

import (
    mg "mingle"
    "mingle/parser"
    "strings"
    "encoding/base64"
    "bitgirder/objpath"
    "fmt"
)

func strToBool( s mg.String, path objpath.PathNode ) ( mg.Value, error ) {
    switch lc := strings.ToLower( string( s ) ); lc { 
    case "true": return mg.Boolean( true ), nil
    case "false": return mg.Boolean( false ), nil
    }
    errTmpl :="Invalid boolean value: %s"
    errStr := mg.QuoteValue( s )
    return nil, mg.NewCastErrorf( path, errTmpl, errStr )
}

func castBoolean( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Boolean: return v, nil
    case mg.String: return strToBool( v, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castBuffer( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Buffer: return v, nil
    case mg.String: 
        buf, err := base64.StdEncoding.DecodeString( string( v ) )
        if err == nil { return mg.Buffer( buf ), nil }
        msg := "Invalid base64 string: %s"
        return nil, mg.NewCastErrorf( path, msg, err.Error() )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castString( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.String: return mgVal, nil
    case mg.Boolean, mg.Int32, mg.Int64, mg.Uint32, mg.Uint64, mg.Float32, 
         mg.Float64:
        return mg.String( v.( fmt.Stringer ).String() ), nil
    case mg.Timestamp: return mg.String( v.Rfc3339Nano() ), nil
    case mg.Buffer:
        b64 := base64.StdEncoding.EncodeToString( []byte( v ) ) 
        return mg.String( b64 ), nil
    case *mg.Enum: return mg.String( v.Value.ExternalForm() ), nil
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func isDecimalNumString( s mg.String ) bool {
    return strings.IndexAny( string( s ), "eE." ) >= 0
}

func parseNumberForCast( 
    s mg.String, 
    numTyp *mg.QualifiedTypeName, 
    path objpath.PathNode ) ( mg.Value, error ) {

    asFloat := mg.IsIntegerTypeName( numTyp ) && isDecimalNumString( s )
    parseTyp := numTyp
    if asFloat { parseTyp = mg.QnameFloat64 }
    val, err := mg.ParseNumber( string( s ), parseTyp )
    if ne, ok := err.( *mg.NumberFormatError ); ok {
        err = mg.NewCastError( path, ne.Error() )
    }
    if err != nil || ( ! asFloat ) { return val, err }
    f64 := float64( val.( mg.Float64 ) )
    switch {
    case numTyp.Equals( mg.QnameInt32 ): val = mg.Int32( int32( f64 ) )
    case numTyp.Equals( mg.QnameUint32 ): val = mg.Uint32( uint32( f64 ) )
    case numTyp.Equals( mg.QnameInt64 ): val = mg.Int64( int64( f64 ) )
    case numTyp.Equals( mg.QnameUint64 ): val = mg.Uint64( uint64( f64 ) )
    }
    return val, nil
}

func castInt32( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return v, nil
    case mg.Int64: return mg.Int32( v ), nil
    case mg.Uint32: return mg.Int32( int32( v ) ), nil
    case mg.Uint64: return mg.Int32( int32( v ) ), nil
    case mg.Float32: return mg.Int32( int32( v ) ), nil
    case mg.Float64: return mg.Int32( int32( v ) ), nil
    case mg.String: return parseNumberForCast( v, mg.QnameInt32, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castInt64( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return mg.Int64( v ), nil
    case mg.Int64: return v, nil
    case mg.Uint32: return mg.Int64( int64( v ) ), nil
    case mg.Uint64: return mg.Int64( int64( v ) ), nil
    case mg.Float32: return mg.Int64( int64( v ) ), nil
    case mg.Float64: return mg.Int64( int64( v ) ), nil
    case mg.String: return parseNumberForCast( v, mg.QnameInt64, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castUint32(
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return mg.Uint32( uint32( v ) ), nil
    case mg.Uint32: return v, nil
    case mg.Int64: return mg.Uint32( uint32( v ) ), nil
    case mg.Uint64: return mg.Uint32( uint32( v ) ), nil
    case mg.Float32: return mg.Uint32( uint32( v ) ), nil
    case mg.Float64: return mg.Uint32( uint32( v ) ), nil
    case mg.String: return parseNumberForCast( v, mg.QnameUint32, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castUint64(
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return mg.Uint64( uint64( v ) ), nil
    case mg.Uint32: return mg.Uint64( uint64( v ) ), nil
    case mg.Int64: return mg.Uint64( uint64( v ) ), nil
    case mg.Uint64: return v, nil
    case mg.Float32: return mg.Uint64( uint64( v ) ), nil
    case mg.Float64: return mg.Uint64( uint64( v ) ), nil
    case mg.String: return parseNumberForCast( v, mg.QnameUint64, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castFloat32( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return mg.Float32( float32( v ) ), nil
    case mg.Int64: return mg.Float32( float32( v ) ), nil
    case mg.Uint32: return mg.Float32( float32( v ) ), nil
    case mg.Uint64: return mg.Float32( float32( v ) ), nil
    case mg.Float32: return v, nil
    case mg.Float64: return mg.Float32( float32( v ) ), nil
    case mg.String: return parseNumberForCast( v, mg.QnameFloat32, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castFloat64( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Int32: return mg.Float64( float64( v ) ), nil
    case mg.Int64: return mg.Float64( float64( v ) ), nil
    case mg.Uint32: return mg.Float64( float64( v ) ), nil
    case mg.Uint64: return mg.Float64( float64( v ) ), nil
    case mg.Float32: return mg.Float64( float64( v ) ), nil
    case mg.Float64: return v, nil
    case mg.String: return parseNumberForCast( v, mg.QnameFloat64, path )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castTimestamp( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case mg.Timestamp: return v, nil
    case mg.String:
        tm, err := parser.ParseTimestamp( string( v ) )
        if err == nil { return tm, nil }
        msg := "Invalid timestamp: %s"
        return nil, mg.NewCastErrorf( path, msg, err.Error() )
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castSymbolMap( 
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    switch v := mgVal.( type ) {
    case *mg.SymbolMap: return v, nil
    }
    return nil, mg.NewTypeCastErrorValue( callTyp, mgVal, path )
}

// switch compares based on qname not at itself since we may be dealing with
// restriction types, meaning that if at is mingle:core@v1/String~"a", it is a
// string (has qname mingle:core@v1/String) but will not equal mg.TypeString
// itself
func castAtomicUnrestricted(
    mgVal mg.Value, 
    at *mg.AtomicTypeReference, 
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( mg.Value, error ) {

    if _, ok := mgVal.( *mg.Null ); ok {
        if at.Equals( mg.TypeNull ) { return mgVal, nil }
        return nil, newNullCastError( path )
    }
    switch nm := at.Name; {
    case nm.Equals( mg.QnameValue ): return mgVal, nil
    case nm.Equals( mg.QnameBoolean ): 
        return castBoolean( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameBuffer ): 
        return castBuffer( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameString ): 
        return castString( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameInt32 ): 
        return castInt32( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameInt64 ): 
        return castInt64( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameUint32 ): 
        return castUint32( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameUint64 ): 
        return castUint64( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameFloat32 ): 
        return castFloat32( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameFloat64 ): 
        return castFloat64( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameTimestamp ): 
        return castTimestamp( mgVal, at, callTyp, path )
    case nm.Equals( mg.QnameSymbolMap ): 
        return castSymbolMap( mgVal, at, callTyp, path )
    }
    return nil, mg.NewTypeCastErrorValue( at, mgVal, path )
}

func checkRestriction( 
    val mg.Value, 
    at *mg.AtomicTypeReference, 
    path objpath.PathNode ) error {

    if at.Restriction.AcceptsValue( val ) { return nil }
    return mg.NewCastErrorf( 
        path, "Value %s does not satisfy restriction %s",
        mg.QuoteValue( val ), at.Restriction.ExternalForm() )
}

func castAtomicWithCallType(
    mgVal mg.Value,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    path objpath.PathNode ) ( val mg.Value, err error ) {

    val, err = castAtomicUnrestricted( mgVal, at, callTyp, path )
    if err == nil && at.Restriction != nil { 
        err = checkRestriction( val, at, path ) 
    }
    return
}

func completeCastEnum(
    id *mg.Identifier, 
    ed *EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    if res := ed.GetValue( id ); res != nil { return res, nil }
    tmpl := "illegal value for enum %s: %s"
    return nil, mg.NewCastErrorf( path, tmpl, ed.GetName(), id )
}

func castEnumFromString( 
    s string, ed *EnumDefinition, path objpath.PathNode ) ( *mg.Enum, error ) {

    id, err := parser.ParseIdentifier( s )
    if err != nil {
        tmpl := "invalid enum value %q: %s"
        return nil, mg.NewCastErrorf( path, tmpl, s, err )
    }
    return completeCastEnum( id, ed, path )
}

func castEnum( 
    val mg.Value, 
    ed *EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    switch v := val.( type ) {
    case mg.String: return castEnumFromString( string( v ), ed, path )
    case *mg.Enum: 
        if v.Type.Equals( ed.GetName() ) {
            return completeCastEnum( v.Value, ed, path )
        }
    }
    t := ed.GetName().AsAtomicType()
    return nil, mg.NewTypeCastErrorValue( t, val, path )
}
