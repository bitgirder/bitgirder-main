package main

import (
    "fmt"
    "mingle/testgen"
    pt "mingle/parser/testing"
    mg "mingle"
)

func ptTyp( nm string ) *mg.QualifiedTypeName {
    str := fmt.Sprintf( "%s/%s", "mingle:parser:testing@v1", nm ) 
    return mg.MustQualifiedTypeName( str )
}

func idListAsValue( ids []pt.Identifier ) *mg.List {
    vals := make( []mg.Value, len( ids ) )
    for i, val := range ids { vals[ i ] = asValue( val ) }
    return mg.NewListValues( vals )
}

func numAsStruct( num *pt.NumericToken ) *mg.Struct {
    str := func( s string ) mg.Value {
        if s == "" { return mg.NullVal }
        return mg.String( s )
    }
    return mg.MustStruct( ptTyp( "NumericToken" ),
        "negative", num.Negative,
        "int", str( num.Int ),
        "frac", str( num.Frac ),
        "exp", str( num.Exp ),
        "exp-char", num.ExpChar,
    )
}

func idAsStruct( id pt.Identifier ) *mg.Struct {
    l := mg.MakeList( len( id ) )
    for i, part := range id { l.Add( mg.String( part ) ) }
    return mg.MustStruct( ptTyp( "Identifier" ), "parts", l )
}

func nsAsStruct( ns *pt.Namespace ) *mg.Struct {
    parts := mg.MakeList( len( ns.Parts ) )
    for i, part := range ns.Parts { parts.Add( asValue( part ) ) }
    return mg.MustStruct( ptTyp( "Namespace" ),
        "parts", parts,
        "version", asValue( ns.Version ),
    )
}

func qnAsStruct( qn *pt.QualifiedTypeName ) *mg.Struct {
    return mg.MustStruct( ptTyp( "QualifiedTypeName" ),
        "namespace", asValue( qn.Namespace ),
        "name", asValue( qn.Name ),
    )
}

func idNmAsStruct( nm *pt.IdentifiedName ) *mg.Struct {
    return mg.MustStruct( ptTyp( "IdentifiedName" ),
        "namespace", asValue( nm.Namespace ),
        "names", idListAsValue( nm.Names ),
    )
}

func asRangeValue( v interface{} ) mg.Value {
    if tm, ok := v.( pt.Timestamp ); ok { v = mg.MustTimestamp( string( tm ) ) }
    return mg.MustValue( v )
}

func rngAsStruct( rr *pt.RangeRestriction ) *mg.Struct {
    return mg.MustStruct( ptTyp( "RangeRestriction" ),
        "min-closed", rr.MinClosed,
        "min", asRangeValue( rr.Min ),
        "max", asRangeValue( rr.Max ),
        "max-closed", rr.MaxClosed,
    )
}

func regxAsStruct( regx pt.RegexRestriction ) *mg.Struct {
    return mg.MustStruct( ptTyp( "RegexRestriction" ), 
        "pattern", string( regx ),
    )
}

func atomicAsStruct( at *pt.AtomicTypeReference ) *mg.Struct {
    return mg.MustStruct( ptTyp( "AtomicTypeReference" ),
        "name", asValue( at.Name ),
        "restriction", asValue( at.Restriction ),
    )
}

func ltAsStruct( lt *pt.ListTypeReference ) *mg.Struct {
    return mg.MustStruct( ptTyp( "ListTypeReference" ),
        "element-type", asValue( lt.ElementType ),
        "allows-empty", lt.AllowsEmpty,
    )
}

func ntAsStruct( nt *pt.NullableTypeReference ) *mg.Struct {
    return mg.MustStruct( ptTyp( "NullableTypeReference" ),
        "type", asValue( nt.Type ),
    )
}

func parseErrAsStruct( pe *pt.ParseErrorExpect ) *mg.Struct {
    return mg.MustStruct( ptTyp( "ParseErrorExpect" ),
        "col", pe.Col,
        "message", pe.Message,
    )
}

func restrictErrAsStruct( re pt.RestrictionErrorExpect ) *mg.Struct {
    return mg.MustStruct( ptTyp( "RestrictionErrorExpect" ),
        "message", string( re ),
    )
}

func asValue( val interface{} ) mg.Value {
    if val == nil { return nil }
    switch v := val.( type ) {
    case nil: return nil
    case pt.StringToken: 
        return mg.MustStruct( ptTyp( "StringToken" ), "string", string( v ) )
    case *pt.NumericToken: return numAsStruct( v )
    case pt.Identifier: return idAsStruct( v )
    case *pt.Namespace: return nsAsStruct( v )
    case pt.DeclaredTypeName:
        return mg.MustStruct( ptTyp( "DeclaredTypeName" ), "name", string( v ) )
    case *pt.QualifiedTypeName: return qnAsStruct( v )
    case *pt.IdentifiedName: return idNmAsStruct( v )
    case *pt.AtomicTypeReference: return atomicAsStruct( v )
    case *pt.RangeRestriction: return rngAsStruct( v )
    case pt.RegexRestriction: return regxAsStruct( v )
    case pt.Timestamp:
        return mg.MustStruct( ptTyp( "Timestamp" ), "time", string( v ) )
    case *pt.ListTypeReference: return ltAsStruct( v )
    case *pt.NullableTypeReference: return ntAsStruct( v )
    case *pt.ParseErrorExpect: return parseErrAsStruct( v )
    case pt.RestrictionErrorExpect: return restrictErrAsStruct( v )
    }
    panic( fmt.Errorf( "unhandled value: %T", val ) )
}

type testData []*pt.CoreParseTest

func ( td testData ) Len() int { return len( td ) }

func ( td testData ) StructAt( i int ) *mg.Struct { 
    t := td[ i ]
    return mg.MustStruct( ptTyp( "CoreParseTest" ),
        "test-type", string( t.TestType ),
        "in", t.In,
        "external-form", t.ExternalForm,
        "expect", asValue( t.Expect ),
        "error", asValue( t.Err ),
    )
}

func main() { testgen.WriteStructFile( testData( pt.CoreParseTests ) ) }
