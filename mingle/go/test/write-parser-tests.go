package main

import (
    "log"
    "fmt"
    "flag"
    "os"
    "path/filepath"
    bgio "bitgirder/io"
    pt "mingle/parser/testing"
)

const (
    fileVersion = int32( 1 )
    
    fldTestType = int8( 1 )
    fldInput = int8( 2 )
    fldExpect = int8( 3 )
    fldError = int8( 4 )
    fldEnd = int8( 5 )
    fldExternalForm = int8( 6 )

    typeIdentifier = int8( 1 )
    typeNamespace = int8( 2 )
    typeDeclaredTypeName = int8( 3 )
    // 4: used to be RelativeTypeName, now removed from mingle
    typeQualifiedTypeName = int8( 5 )
    typeIdentifiedName = int8( 6 )
    typeRegexRestriction = int8( 7 )
    typeRangeRestriction = int8( 8 )
    typeAtomicTypeReference = int8( 9 )
    typeListTypeReference = int8( 10 )
    typeNullableTypeReference = int8( 11 )
    typeNil = int8( 12 )
    typeInt32 = int8( 13 )
    typeInt64 = int8( 14 )
    typeFloat32 = int8( 15 )
    typeFloat64 = int8( 16 )
    typeString = int8( 17 )
    typeTimestamp = int8( 18 )
    typeBoolean = int8( 19 )
    typeParseError = int8( 20 )
    typeRestrictionError = int8( 21 )
    typeStringToken = int8( 22 )
    typeNumericToken = int8( 23 )
    typeUint32 = int8( 24 )
    typeUint64 = int8( 25 )

    eltTypeFileEnd = int8( 0 )
    eltTypeParseTest = int8( 1 )
)

var outFile string

func init() {
    flag.StringVar( &outFile, "out-file", "", "Dest file for test data" )
}

func initOutFile() ( wr *bgio.BinWriter, err error ) {
    if dir := filepath.Dir( outFile ); dir != "." {
        if err = os.MkdirAll( dir, os.FileMode( 0755 ) ); err != nil {
            err = fmt.Errorf( "Couldn't create parent dir %s: %s", dir, err )
            return
        }
    }
    var f *os.File
    if f, err = os.Create( outFile ); err != nil {
        err = fmt.Errorf( "Couldn't create %s: %s", outFile, err )
        return
    } else { wr = bgio.NewLeWriter( f ) }
    return 
}

func writeNilVal( wr *bgio.BinWriter ) error {
    return writeTypeCode( typeNil, wr )
}

func implWritePrimValue( tc int8, val interface{}, wr *bgio.BinWriter ) error {
    if err := writeTypeCode( tc, wr ); err != nil { return err }
    switch v := val.( type ) {
    case string: return wr.WriteUtf8( v )
    case bool:
        i := int8( 0 )
        if v { i = int8( 1 ) }
        return wr.WriteInt8( i )
    }
    return wr.WriteBin( val )
}

func writePrimValue( val interface{}, wr *bgio.BinWriter ) ( err error ) {
    switch v := val.( type ) {
    case nil: err = writeNilVal( wr )
    case bool: err = implWritePrimValue( typeBoolean, v, wr )
    case int32: err = implWritePrimValue( typeInt32, v, wr )
    case int64: err = implWritePrimValue( typeInt64, v, wr )
    case uint32: err = implWritePrimValue( typeUint32, v, wr )
    case uint64: err = implWritePrimValue( typeUint64, v, wr )
    case float32: err = implWritePrimValue( typeFloat32, v, wr )
    case float64: err = implWritePrimValue( typeFloat64, v, wr )
    case string: err = implWritePrimValue( typeString, v, wr )
    case pt.Timestamp: 
        err = implWritePrimValue( typeTimestamp, string( v ), wr )
    default: panic( fmt.Errorf( "Unhandled value type: %T", val ) )
    }
    return 
}

func writeHeader( wr *bgio.BinWriter ) ( err error ) {
    return wr.WriteInt32( fileVersion )
}

func writeFieldCode( fld int8, wr *bgio.BinWriter ) error {
    return wr.WriteInt8( fld )
}

func writeTypeCode( tc int8, wr *bgio.BinWriter ) error {
    return wr.WriteInt8( tc )
}

func writeStringToken( s pt.StringToken, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeStringToken, wr ); err != nil { return }
    return wr.WriteUtf8( string( s ) )
}

func writeNumericToken( n *pt.NumericToken, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeNumericToken, wr ); err != nil { return }
    if err = writePrimValue( n.Negative, wr ); err != nil { return }
    for _, s := range []string{ n.Int, n.Frac, n.Exp, n.ExpChar } {
        if err = wr.WriteUtf8( s ); err != nil { return }
    }
    return nil
}

func writeIdentifier( id pt.Identifier, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeIdentifier, wr ); err != nil { return }
    if err = wr.WriteInt32( int32( len( id ) ) ); err != nil { return }
    for _, part := range id {
        if err = wr.WriteUtf8( string( part ) ); err != nil { return }
    }
    return
}

func writeIdentifiers( ids []pt.Identifier, wr *bgio.BinWriter ) ( err error ) {
    if err = wr.WriteInt32( int32( len( ids ) ) ); err != nil { return }
    for _, part := range ids {
        if err = writeIdentifier( part, wr ); err != nil { return }
    }
    return
}

func writeNamespace( ns *pt.Namespace, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeNamespace, wr ); err != nil { return }
    if err = writeIdentifiers( ns.Parts, wr ); err != nil { return }
    if err = writeIdentifier( ns.Version, wr ); err != nil { return }
    return
}

func writeDeclaredTypeName( 
    nm pt.DeclaredTypeName, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeDeclaredTypeName, wr ); err != nil { return }
    if err = wr.WriteUtf8( string( nm ) ); err != nil { return }
    return
}

func writeQualifiedTypeName( 
    qn *pt.QualifiedTypeName, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeQualifiedTypeName, wr ); err != nil { return }
    if err = writeNamespace( qn.Namespace, wr ); err != nil { return }
    if err = writeDeclaredTypeName( qn.Name, wr ); err != nil { return }
    return
}

func writeIdentifiedName( 
    nm *pt.IdentifiedName, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeIdentifiedName, wr ); err != nil { return }
    if err = writeNamespace( nm.Namespace, wr ); err != nil { return }
    if err = writeIdentifiers( nm.Names, wr ); err != nil { return }
    return
}

func writeRegexRestriction( 
    rr pt.RegexRestriction, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeRegexRestriction, wr ); err != nil { return }
    if err = wr.WriteUtf8( string( rr ) ); err != nil { return }
    return
}

func writeRangeRestriction(
    rr *pt.RangeRestriction, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeRangeRestriction, wr ); err != nil { return }
    if err = writePrimValue( rr.MinClosed, wr ); err != nil { return }
    if err = writePrimValue( rr.Min, wr ); err != nil { return }
    if err = writePrimValue( rr.Max, wr ); err != nil { return }
    if err = writePrimValue( rr.MaxClosed, wr ); err != nil { return }
    return
}

func writeRestriction( rVal interface{}, wr *bgio.BinWriter ) error {
    switch v := rVal.( type ) {
    case pt.RegexRestriction: return writeRegexRestriction( v, wr )
    case *pt.RangeRestriction: return writeRangeRestriction( v, wr )
    case nil: return writeNilVal( wr )
    }
    return fmt.Errorf( "Unhandled restriction type: %T", rVal )
}

func writeAtomicTypeReference(
    at *pt.AtomicTypeReference, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeAtomicTypeReference, wr ); err != nil { return }
    if err = writeValue( at.Name, wr ); err != nil { return }
    if err = writeRestriction( at.Restriction, wr ); err != nil { return }
    return
}

func writeListTypeReference( 
    lt *pt.ListTypeReference, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeListTypeReference, wr ); err != nil { return }
    if err = writeTypeReference( lt.ElementType, wr ); err != nil { return }
    if err = writePrimValue( lt.AllowsEmpty, wr ); err != nil { return }
    return
}

func writeNullableTypeReference(
    lt *pt.NullableTypeReference, wr *bgio.BinWriter ) ( err error ) {
    if err = writeTypeCode( typeNullableTypeReference, wr ); err != nil {
        return
    }
    if err = writeTypeReference( lt.Type, wr ); err != nil { return }
    return
}

func writeTypeReference( tVal interface{}, wr *bgio.BinWriter ) error {
    switch t := tVal.( type ) {
    case *pt.AtomicTypeReference: return writeAtomicTypeReference( t, wr )
    case *pt.ListTypeReference: return writeListTypeReference( t, wr )
    case *pt.NullableTypeReference: return writeNullableTypeReference( t, wr )
    }
    return fmt.Errorf( "Unhandled type reference: %T", tVal )
}

func writeValue( expct interface{}, wr *bgio.BinWriter ) ( err error ) {
    switch v := expct.( type ) {
    case pt.StringToken: err = writeStringToken( v, wr )
    case *pt.NumericToken: err = writeNumericToken( v, wr )
    case pt.Identifier: err = writeIdentifier( v, wr )
    case *pt.Namespace: err = writeNamespace( v, wr )
    case pt.DeclaredTypeName: err = writeDeclaredTypeName( v, wr )
    case *pt.QualifiedTypeName: err = writeQualifiedTypeName( v, wr )
    case *pt.IdentifiedName: err = writeIdentifiedName( v, wr )
    case *pt.AtomicTypeReference,
         *pt.ListTypeReference,
         *pt.NullableTypeReference:
        err = writeTypeReference( v, wr )
    default: panic( fmt.Errorf( "Unhandled expect val: %T", expct ) )
    }
    return
}

func writeError( ee pt.ErrorExpect, wr *bgio.BinWriter ) ( err error ) {
    switch v := ee.( type ) {
    case *pt.ParseErrorExpect:
        if err = writeTypeCode( typeParseError, wr ); err != nil { return }
        if err = wr.WriteInt32( int32( v.Col ) ); err != nil { return }
        if err = wr.WriteUtf8( v.Message ); err != nil { return }
    case pt.RestrictionErrorExpect:
        if err = writeTypeCode( typeRestrictionError, wr ); err != nil { 
            return
        }
        if err = wr.WriteUtf8( string( v ) ); err != nil { return }
    default: err = fmt.Errorf( "Unhandled error expectation: %T", ee )
    }
    return
}

func writeTest( cpt *pt.CoreParseTest, wr *bgio.BinWriter ) ( err error ) {
    if err = wr.WriteInt8( eltTypeParseTest ); err != nil { return }
    if err = writeFieldCode( fldTestType, wr ); err != nil { return }
    if err = wr.WriteUtf8( string( cpt.TestType ) ); err != nil { return }
    if err = writeFieldCode( fldInput, wr ); err != nil { return }
    if err = wr.WriteUtf8( cpt.In ); err != nil { return }
    if err = writeFieldCode( fldExternalForm, wr ); err != nil { return }
    if err = wr.WriteUtf8( cpt.ExternalForm ); err != nil { return }
    if cpt.Expect != nil {
        if err = writeFieldCode( fldExpect, wr ); err != nil { return }
        if err = writeValue( cpt.Expect, wr ); err != nil { return }
    }
    if cpt.Err != nil {
        if err = writeFieldCode( fldError, wr ); err != nil { return }
        if err = writeError( cpt.Err, wr ); err != nil { return }
    }
    if err = writeFieldCode( fldEnd, wr ); err != nil { return }
    return
}

func writeTrailer( wr *bgio.BinWriter ) error {
    return wr.WriteInt8( eltTypeFileEnd )
}

func writeTests( wr *bgio.BinWriter ) ( err error ) {
    log.Printf( "Writing %s", outFile )
    if err = writeHeader( wr ); err != nil { return }
    for _, cpt := range pt.CoreParseTests {
        if err = writeTest( cpt, wr ); err != nil { return }
    }
    if err = writeTrailer( wr ); err != nil { return }
    return
}

func main() {
    flag.Parse()
    if outFile == "" { log.Fatalf( "Missing output file" ) }
    var err error
    var wr *bgio.BinWriter
    if wr, err = initOutFile(); err == nil { 
        defer wr.Close()
        err = writeTests( wr ) 
    }
    if err != nil { log.Fatal( err ) }
}
