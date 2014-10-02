package main

import (
    "fmt"
    "mingle/testgen"
    "mingle/parser"
    mg "mingle"
    "bytes"
    "bitgirder/objpath"
)

var ( 
    mustStruct = parser.MustStruct
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
)

func ptTyp( val interface{} ) *mg.QualifiedTypeName {
    switch v := val.( type ) {
    case *mg.QualifiedTypeName: return v
    case string: return mkQn( "mingle:parser@v1/" + v )
    }
    panic( fmt.Errorf( "unhandled input to ptTyp: %T", val ) )
}

func idListAsValue( ids []mg.Identifier ) *mg.List {
    vals := make( []mg.Value, len( ids ) )
    for i, val := range ids { vals[ i ] = asValue( val ) }
    return mg.NewListValues( vals )
}

func numAsStruct( num *parser.NumericToken ) *mg.Struct {
    str := func( s string ) mg.Value {
        if s == "" { return mg.NullVal }
        return mg.String( s )
    }
    return mustStruct( ptTyp( "NumericToken" ),
        "int", str( num.Int ),
        "frac", str( num.Frac ),
        "exp", str( num.Exp ),
        "exp-rune", num.ExpRune,
    )
}

func asBytes( val interface{} ) []byte {
    bb := &bytes.Buffer{}
    w := mg.NewWriter( bb )
    var err error
    switch v := val.( type ) {
    case *mg.Identifier: err = w.WriteIdentifier( v )
    case *mg.Namespace: err = w.WriteNamespace( v )
    case *mg.QualifiedTypeName: err = w.WriteQualifiedTypeName( v )
    case *mg.DeclaredTypeName: err = w.WriteDeclaredTypeName( v )
    case mg.TypeReference: err = w.WriteTypeReference( v )
    default: panic( fmt.Errorf( "unhandled type: %T", val ) )
    }
    if err == nil { return bb.Bytes() }
    panic( err )
}

func makeBufferStruct( typ interface{}, val interface{} ) *mg.Struct {
    return mustStruct( ptTyp( typ ), "buffer", asBytes( val ) )
}

func idAsStruct( id *mg.Identifier ) *mg.Struct {
    return makeBufferStruct( mg.QnameIdentifier, id )
}

func nsAsStruct( ns *mg.Namespace ) *mg.Struct {
    return makeBufferStruct( mg.QnameNamespace, ns )
}

func declNmAsStruct( dn *mg.DeclaredTypeName ) *mg.Struct {
    return makeBufferStruct( mkQn( "mingle:core@v1/DeclaredTypeName" ), dn )
}

func qnAsStruct( qn *mg.QualifiedTypeName ) *mg.Struct {
    return makeBufferStruct( mkQn( "mingle:core@v1/QualifiedTypeName" ), qn )
}

func parseErrAsStruct( pe *parser.ParseErrorExpect ) *mg.Struct {
    return mustStruct( ptTyp( "ParseErrorExpect" ),
        "col", pe.Col,
        "message", pe.Message,
    )
}

func restrictErrAsStruct( re parser.RestrictionErrorExpect ) *mg.Struct {
    return mustStruct( ptTyp( "RestrictionErrorExpect" ),
        "message", string( re ),
    )
}

type idPathVisitor struct { l *mg.List }

func ( v idPathVisitor ) Descend( elt interface{} ) error {
    v.l.AddUnsafe( mg.Buffer( asBytes( elt.( *mg.Identifier ) ) ) )
    return nil
}

func ( v idPathVisitor ) List( idx uint64 ) error {
    v.l.AddUnsafe( mg.Uint64( idx ) )
    return nil
}

func idPathAsStruct( path objpath.PathNode ) *mg.Struct {
    v := idPathVisitor{ mg.NewList( mg.TypeOpaqueList ) }
    if err := objpath.Visit( path, v ); err != nil { panic( err ) }
    return mustStruct( mg.QnameIdentifierPath, "path", v.l )
}

func atomicTypeExpressionAsStruct( e *parser.AtomicTypeExpression ) *mg.Struct {
    return mustStruct( ptTyp( "AtomicTypeExpression" ),
        "name", asValue( e.Name ),
        "loc", asValue( e.NameLoc ),
        "restriction", asValue( e.Restriction() ),
    )
}

func ptrTypeExpressionAsStruct( e *parser.PointerTypeExpression ) *mg.Struct {
    return mustStruct( ptTyp( "PointerTypeExpression" ),
        "expression", asValue( e.Expression ),
        "loc", asValue( e.Loc ),
    )
}

func nullableTypeExpressionAsStruct( 
    e *parser.NullableTypeExpression ) *mg.Struct {

    return mustStruct( ptTyp( "NullableTypeExpression" ),
        "expression", asValue( e.Expression ),
        "loc", asValue( e.Loc ),
    )
}

func listTypeExpressionAsStruct( e *parser.ListTypeExpression ) *mg.Struct {
    return mustStruct( ptTyp( "ListTypeExpression" ),
        "expression", asValue( e.Expression ),
        "loc", asValue( e.Loc ),
        "allows-empty", e.AllowsEmpty,
    )
}

func ctRefAsStruct( ctr *parser.CompletableTypeReference ) *mg.Struct {
    return mustStruct( ptTyp( "CompletableTypeReference" ),
        "expression", asValue( ctr.Expression ),
    )
}

func locAsStruct( lc *parser.Location ) *mg.Struct {
    return mustStruct( ptTyp( "Location" ),
        "line", lc.Line,
        "col", lc.Col,
        "source", lc.Source,
    )
}

func rangeRestrictionSyntaxAsStruct( 
    rr *parser.RangeRestrictionSyntax ) *mg.Struct {

    return mustStruct( ptTyp( "RangeRestrictionSyntax" ),
        "loc", asValue( rr.Loc ),
        "left-closed", rr.LeftClosed,
        "left", asValue( rr.Left ),
        "right", asValue( rr.Right ),
        "right-closed", rr.RightClosed,
    )
}

func regexRestrictionSyntaxAsStruct( 
    rr *parser.RegexRestrictionSyntax ) *mg.Struct {
    
    return mustStruct( ptTyp( "RegexRestrictionSyntax" ),
        "pat", rr.Pat,
        "loc", asValue( rr.Loc ),
    )
}

func asValue( val interface{} ) mg.Value {
    if val == nil { return nil }
    switch v := val.( type ) {
    case nil: return nil
    case parser.StringToken: 
        return mustStruct( ptTyp( "StringToken" ), "string", string( v ) )
    case *parser.NumericToken: return numAsStruct( v )
    case *mg.Identifier: return idAsStruct( v )
    case *mg.Namespace: return nsAsStruct( v )
    case *mg.DeclaredTypeName: return declNmAsStruct( v )
    case *mg.QualifiedTypeName: return qnAsStruct( v )
    case *parser.ParseErrorExpect: return parseErrAsStruct( v )
    case parser.RestrictionErrorExpect: return restrictErrAsStruct( v )
    case *parser.CompletableTypeReference: return ctRefAsStruct( v )
    case *parser.AtomicTypeExpression: return atomicTypeExpressionAsStruct( v )
    case *parser.RegexRestrictionSyntax:
        return regexRestrictionSyntaxAsStruct( v )
    case *parser.RangeRestrictionSyntax:
        return rangeRestrictionSyntaxAsStruct( v )
    case *parser.ListTypeExpression: return listTypeExpressionAsStruct( v )
    case *parser.PointerTypeExpression: return ptrTypeExpressionAsStruct( v )
    case *parser.NullableTypeExpression:
        return nullableTypeExpressionAsStruct( v )
    case *parser.Location: return locAsStruct( v )
    case objpath.PathNode: return idPathAsStruct( v )
    }
    panic( fmt.Errorf( "unhandled value: %T", val ) )
}

type testData []*parser.CoreParseTest

func ( td testData ) Len() int { return len( td ) }

func ( td testData ) StructAt( i int ) *mg.Struct { 
    t := td[ i ]
    return mustStruct( ptTyp( "CoreParseTest" ),
        "test-type", string( t.TestType ),
        "in", t.In,
        "external-form", t.ExternalForm,
        "expect", asValue( t.Expect ),
        "error", asValue( t.Err ),
    )
}

func main() { testgen.WriteStructFile( testData( parser.CoreParseTests ) ) }
