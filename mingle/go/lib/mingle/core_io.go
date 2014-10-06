package mingle

import (
    "fmt"
    "bytes"
    "io"
    "time"
//    "log"
    bgio "bitgirder/io"
)

type IoTypeCode uint8

const (
    IoTypeCodeNull = IoTypeCode( uint8( 0x00 ) )
    IoTypeCodeId = IoTypeCode( uint8( 0x01 ) )
    IoTypeCodeNs = IoTypeCode( uint8( 0x02 ) )
    IoTypeCodeDeclNm = IoTypeCode( uint8( 0x03 ) )
    IoTypeCodeQn = IoTypeCode( uint8( 0x04 ) )
    IoTypeCodeAtomTyp = IoTypeCode( uint8( 0x05 ) )
    IoTypeCodeListTyp = IoTypeCode( uint8( 0x06 ) )
    IoTypeCodeNullableTyp = IoTypeCode( uint8( 0x07 ) )
    IoTypeCodePointerTyp = IoTypeCode( uint8( 0x08 ) )
    IoTypeCodeRegexRestrict = IoTypeCode( uint8( 0x09 ) )
    IoTypeCodeRangeRestrict = IoTypeCode( uint8( 0x0a ) )
    IoTypeCodeBool = IoTypeCode( uint8( 0x0b ) )
    IoTypeCodeString = IoTypeCode( uint8( 0x0c ) )
    IoTypeCodeInt32 = IoTypeCode( uint8( 0x0d ) )
    IoTypeCodeInt64 = IoTypeCode( uint8( 0x0e ) )
    IoTypeCodeUint32 = IoTypeCode( uint8( 0x0f ) )
    IoTypeCodeUint64 = IoTypeCode( uint8( 0x10 ) )
    IoTypeCodeFloat32 = IoTypeCode( uint8( 0x11 ) )
    IoTypeCodeFloat64 = IoTypeCode( uint8( 0x12 ) )
    IoTypeCodeTimestamp = IoTypeCode( uint8( 0x13 ) )
    IoTypeCodeBuffer = IoTypeCode( uint8( 0x14 ) )
    IoTypeCodeEnum = IoTypeCode( uint8( 0x15 ) )
    IoTypeCodeSymMap = IoTypeCode( uint8( 0x16 ) )
    IoTypeCodeField = IoTypeCode( uint8( 0x17 ) )
    IoTypeCodeStruct = IoTypeCode( uint8( 0x18 ) )
    IoTypeCodeList = IoTypeCode( uint8( 0x19 ) )
    IoTypeCodeEnd = IoTypeCode( uint8( 0x1a ) )
)

type BinIoError struct { msg string }

func ( e *BinIoError ) Error() string { return e.msg }

func NewBinIoErrorOffset( off int64, msg string ) *BinIoError {
    return &BinIoError{ fmt.Sprintf( "[offset %d]: %s", off, msg ) }
} 

type BinWriter struct { *bgio.BinWriter }

func AsWriter( w *bgio.BinWriter ) *BinWriter { return &BinWriter{ w } }

func NewWriter( w io.Writer ) *BinWriter { 
    return AsWriter( bgio.NewLeWriter( w ) )
}

// helper to power public funcs that convert a value to []byte
func writeAsBytes( f func ( *BinWriter ) error ) []byte {
    buf := &bytes.Buffer{}
    w := NewWriter( buf )
    if err := f( w ); err != nil {
        panic( libErrorf( "unhandled error converting to byte: %s", err ) )
    }
    return buf.Bytes()
}

func ( w *BinWriter ) WriteTypeCode( tc IoTypeCode ) error {
    return w.WriteUint8( uint8( tc ) )
}

func ( w *BinWriter ) WriteNull() error { 
    return w.WriteTypeCode( IoTypeCodeNull ) 
}

func ( w *BinWriter ) writeBool( b bool ) ( err error ) {
    return w.WriteBool( b )
}

func ( w *BinWriter ) WriteIdentifier( id *Identifier ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeId ); err != nil { return }
    if err = w.WriteUint8( uint8( len( id.parts ) ) ); err != nil { return }
    for _, part := range id.parts {
        if err = w.WriteUtf8( part ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) WriteIdentifiers( ids []*Identifier ) ( err error ) {
    if err = w.WriteUint8( uint8( len( ids ) ) ); err != nil { return }
    for _, id := range ids {
        if err = w.WriteIdentifier( id ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) WriteNamespace( ns *Namespace ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeNs ); err != nil { return }
    if err = w.WriteIdentifiers( ns.Parts ); err != nil { return }
    return w.WriteIdentifier( ns.Version )
}

func ( w *BinWriter ) WriteDeclaredTypeName( 
    n *DeclaredTypeName ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeDeclNm ); err != nil { return }
    return w.WriteUtf8( n.nm )
}

func ( w *BinWriter ) WriteQualifiedTypeName( 
    qn *QualifiedTypeName ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeQn ); err != nil { return }
    if err = w.WriteNamespace( qn.Namespace ); err != nil { return }
    return w.WriteDeclaredTypeName( qn.Name )
}

func ( w *BinWriter ) WriteTypeName( nm TypeName ) error {
    switch v := nm.( type ) {
    case *DeclaredTypeName: return w.WriteDeclaredTypeName( v )
    case *QualifiedTypeName: return w.WriteQualifiedTypeName( v )
    }
    panic( libErrorf( "unhandled type name: %T", nm ) )
}

func ( w *BinWriter ) writeRegexRestriction( 
    rr *RegexRestriction ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeRegexRestrict ); err != nil { return }
    return w.WriteUtf8( rr.src )
}

func ( w *BinWriter ) writeEnum( en *Enum ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeEnum ); err != nil { return }
    if err = w.WriteQualifiedTypeName( en.Type ); err != nil { return }
    if err = w.WriteIdentifier( en.Value ); err != nil { return }
    return
}

func ( w *BinWriter ) WriteScalarValue( val Value ) error {
    switch v := val.( type ) {
    case nil: return w.WriteNull()
    case *Null: return w.WriteNull()
    case Boolean: 
        if err := w.WriteTypeCode( IoTypeCodeBool ); err != nil { return err }
        return w.WriteBool( bool( v ) )
    case Buffer:
        if err := w.WriteTypeCode( IoTypeCodeBuffer ); err != nil { return err }
        return w.WriteBuffer32( []byte( v ) )
    case String:
        if err := w.WriteTypeCode( IoTypeCodeString ); err != nil { return err }
        return w.WriteUtf8( string( v ) )
    case Int32:
        if err := w.WriteTypeCode( IoTypeCodeInt32 ); err != nil { return err }
        return w.WriteInt32( int32( v ) )
    case Int64:
        if err := w.WriteTypeCode( IoTypeCodeInt64 ); err != nil { return err }
        return w.WriteInt64( int64( v ) )
    case Uint32:
        if err := w.WriteTypeCode( IoTypeCodeUint32 ); err != nil { return err }
        return w.WriteUint32( uint32( v ) )
    case Uint64:
        if err := w.WriteTypeCode( IoTypeCodeUint64 ); err != nil { return err }
        return w.WriteUint64( uint64( v ) )
    case Float32:
        if err := w.WriteTypeCode( IoTypeCodeFloat32 ); err != nil { 
            return err 
        }
        return w.WriteFloat32( float32( v ) )
    case Float64:
        if err := w.WriteTypeCode( IoTypeCodeFloat64 ); err != nil { 
            return err 
        }
        return w.WriteFloat64( float64( v ) )
    case Timestamp:
        if err := w.WriteTypeCode( IoTypeCodeTimestamp ); err != nil { 
            return err 
        }
        if err := w.WriteInt64( time.Time( v ).Unix() ); err != nil { 
            return err
        }
        return w.WriteInt32( int32( time.Time( v ).Nanosecond() ) )
    case *Enum: return w.writeEnum( v )
    }
    panic( libErrorf( "unhandled value: %T", val ) )
}

func ( w *BinWriter ) writeRangeValue( val Value ) error {
    switch val.( type ) {
    case nil, Null, String, Int32, Int64, Uint32, Uint64, Float32, Float64,
         Timestamp:
        return w.WriteScalarValue( val )
    }
    panic( libErrorf( "unhandled range val: %T", val ) )
}

func ( w *BinWriter ) writeRangeRestriction( 
    rr *RangeRestriction ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeRangeRestrict ); err != nil { return }
    if err = w.writeBool( rr.MinClosed() ); err != nil { return }
    if err = w.writeRangeValue( rr.Min() ); err != nil { return }
    if err = w.writeRangeValue( rr.Max() ); err != nil { return }
    return w.writeBool( rr.MaxClosed() )
}

func ( w *BinWriter ) WriteAtomicTypeReference( 
    at *AtomicTypeReference ) ( err error ) {

    if err = w.WriteTypeCode( IoTypeCodeAtomTyp ); err != nil { return }
    if err = w.WriteTypeName( at.Name() ); err != nil { return }
    switch r := at.Restriction().( type ) {
    case nil: return w.WriteNull()
    case *RegexRestriction: return w.writeRegexRestriction( r )
    case *RangeRestriction: return w.writeRangeRestriction( r )
    default: panic( libErrorf( "unhandled restriction: %T", r ) )
    }
    return
}

func ( w *BinWriter ) WriteListTypeReference( 
    lt *ListTypeReference ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeListTyp ); err != nil { return }
    if err = w.WriteTypeReference( lt.ElementType ); err != nil { return }
    return w.writeBool( lt.AllowsEmpty )
}

func ( w *BinWriter ) WriteNullableTypeReference( 
    nt *NullableTypeReference ) ( err error ) {
    if err = w.WriteTypeCode( IoTypeCodeNullableTyp ); err != nil { return }
    return w.WriteTypeReference( nt.Type )
}

func ( w *BinWriter ) WritePointerTypeReference(
    pt *PointerTypeReference ) error {

    if err := w.WriteTypeCode( IoTypeCodePointerTyp ); err != nil { return err }
    return w.WriteTypeReference( pt.Type )
}

func ( w *BinWriter ) WriteTypeReference( typ TypeReference ) error {
    switch v := typ.( type ) {
    case *AtomicTypeReference: return w.WriteAtomicTypeReference( v )
    case *ListTypeReference: return w.WriteListTypeReference( v )
    case *NullableTypeReference: return w.WriteNullableTypeReference( v )
    case *PointerTypeReference: return w.WritePointerTypeReference( v )
    }
    panic( libErrorf( "unhandled type reference: %T", typ ) )
}

func TypeReferenceAsBytes( typ TypeReference ) []byte {
    return writeAsBytes( func( w *BinWriter ) error { 
        return w.WriteTypeReference( typ )
    })
}

func IdentifierAsBytes( id *Identifier ) []byte {
    return writeAsBytes( func( w *BinWriter ) error {
        return w.WriteIdentifier( id )
    })
}

func QualifiedTypeNameAsBytes( qn *QualifiedTypeName ) []byte {
    return writeAsBytes( func( w *BinWriter ) error {
        return w.WriteQualifiedTypeName( qn )
    })
}

func NamespaceAsBytes( ns *Namespace ) []byte {
    return writeAsBytes( func( w *BinWriter ) error {
        return w.WriteNamespace( ns )
    })
}

type offsetTracker struct {
    rd io.Reader
    off int64
    saved int16
    forUnread int16
}

func ( ot *offsetTracker ) Read( p []byte ) ( int, error ) {
    resAdd := 0
    ot.forUnread = -1
    if ot.saved >= 0 && len( p ) > 0 {
        p[ 0 ] = byte( ot.saved )
        ot.saved = -1
        resAdd = 1
    }
    res := 0
    var err error
    if len( p ) > 0 { res, err = ot.rd.Read( p[ resAdd : ] ) }
    if err == nil { 
        res += resAdd 
        ot.forUnread = int16( p[ res - 1 ] )
    }
    ot.off += int64( res )
    return res, err
}

func ( ot *offsetTracker ) UnreadByte() error {
    if ot.forUnread < 0 { return libErrorf( "Nothing to Unread()" ) }
    ot.off -= 1
    ot.saved, ot.forUnread = ot.forUnread, -1
    return nil
}

type BinReader struct {
    ot *offsetTracker
    *bgio.BinReader
}

func NewReader( r io.Reader ) *BinReader {
    ot := &offsetTracker{ rd: r, off: 0, saved: -1 }
    return &BinReader{ ot: ot, BinReader: bgio.NewLeReader( ot ) }
}

func NewReaderBytes( buf []byte ) *BinReader {
    return NewReader( bytes.NewBuffer( buf ) )
}

func ( r *BinReader ) offset() int64 {
    return r.ot.off
}

func ( r *BinReader ) IoErrorf( tmpl string, args ...interface{} ) *BinIoError {
    return NewBinIoErrorOffset( r.offset() - 1, fmt.Sprintf( tmpl, args... ) )
}

func ( r *BinReader ) ReadTypeCode() ( IoTypeCode, error ) {
    res, err := r.ReadUint8()
    return IoTypeCode( res ), err
}

// State of reader is undefined after a call to this method that returns a
// non-nil error
func ( r *BinReader ) PeekTypeCode() ( IoTypeCode, error ) {
    res, err := r.ReadTypeCode()
    if err2 := r.ot.UnreadByte(); err2 != nil { panic( err2 ) }
    return res, err
}

func ( r *BinReader ) ExpectTypeCode( expct IoTypeCode ) ( IoTypeCode, error ) {
    res, err := r.ReadTypeCode()
    if err != nil { return 0, err }
    if res == expct { return res, nil }
    tmpl := "Expected type code 0x%02x but got 0x%02x"
    return 0, r.IoErrorf( tmpl, expct, res )
}

func ( r *BinReader ) readBool() ( bool, error ) { return r.ReadBool() }

func ( r *BinReader ) readIdPart() ( string, error ) {
    off := r.offset()
    s, err := r.ReadUtf8()
    if err != nil { return "", err }
    if err := getIdentifierPartError( s ); err != nil { 
        return "", NewBinIoErrorOffset( off, err.Error() )
    }
    return s, nil
}

func ( r *BinReader ) ReadIdentifier() ( id *Identifier, err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeId ); err != nil { return }
    var sz uint8
    if sz, err = r.ReadUint8(); err != nil { return }
    parts := make( []string, sz )
    for i := uint8( 0 ); i < sz; i++ {
        if parts[ i ], err = r.readIdPart(); err != nil { return }
    }
    return NewIdentifierUnsafe( parts ), nil
}

func ( r *BinReader ) ReadIdentifiers() ( ids []*Identifier, err error ) {
    var sz uint8
    if sz, err = r.ReadUint8(); err != nil { return }
    ids = make( []*Identifier, sz )
    for i := uint8( 0 ); i < sz; i++ {
        if ids[ i ], err = r.ReadIdentifier(); err != nil { return }
    }
    return
}

func ( r *BinReader ) ReadNamespace() ( ns *Namespace, err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeNs ); err != nil { return }
    ns = &Namespace{}
    if ns.Parts, err = r.ReadIdentifiers(); err != nil { return }
    if ns.Version, err = r.ReadIdentifier(); err != nil { return }
    return
}

func ( r *BinReader ) ReadDeclaredTypeName() ( nm *DeclaredTypeName,    
                                               err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeDeclNm ); err != nil { return }
    var s string
    off := r.offset()
    if s, err = r.ReadUtf8(); err != nil { return }
    if nm, err = CreateDeclaredTypeName( s ); err != nil {
        err = NewBinIoErrorOffset( off, err.Error() )
    }
    return
}

func ( r *BinReader ) ReadQualifiedTypeName() ( qn *QualifiedTypeName,
                                                err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeQn ); err != nil { return }
    qn = &QualifiedTypeName{}
    if qn.Namespace, err = r.ReadNamespace(); err != nil { return }
    if qn.Name, err = r.ReadDeclaredTypeName(); err != nil { return }
    return
}

func ( r *BinReader ) ReadTypeName() ( nm TypeName, err error ) {
    var tc IoTypeCode
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case IoTypeCodeDeclNm: return r.ReadDeclaredTypeName()
    case IoTypeCodeQn: return r.ReadQualifiedTypeName()
    }
    return nil, fmt.Errorf( "mingle: Unrecognized type name code: 0x%02x", tc )
}

func ( r *BinReader ) readEnum() ( en *Enum, err error ) {
    en = &Enum{}
    if en.Type, err = r.ReadQualifiedTypeName(); err != nil { return }
    if en.Value, err = r.ReadIdentifier(); err != nil { return }
    return
}

// tc is already read when this is called
func ( r *BinReader ) ReadScalarValue( tc IoTypeCode ) ( Value, error ) {
    switch tc {
    case IoTypeCodeNull: return NullVal, nil
    case IoTypeCodeString: 
        if s, err := r.ReadUtf8(); err == nil { 
            return String( s ), nil
        } else { return nil, err }
    case IoTypeCodeBuffer:
        if buf, err := r.ReadBuffer32(); err == nil { 
            return Buffer( buf ), nil
        } else { return nil, err }
    case IoTypeCodeTimestamp:
        if ux, err := r.ReadInt64(); err == nil { 
            if ns, err := r.ReadInt32(); err == nil { 
                return Timestamp( time.Unix( ux, int64( ns ) ) ), nil
            } else { return nil, err } 
        } else { return nil, err }
    case IoTypeCodeInt32:
        if i, err := r.ReadInt32(); err == nil {
            return Int32( i ), nil
        } else { return nil, err }
    case IoTypeCodeInt64:
        if i, err := r.ReadInt64(); err == nil { 
            return Int64( i ), nil
        } else { return nil, err }
    case IoTypeCodeUint32:
        if i, err := r.ReadUint32(); err == nil { 
            return Uint32( i ), nil
        } else { return nil, err }
    case IoTypeCodeUint64:
        if i, err := r.ReadUint64(); err == nil { 
            return Uint64( i ), nil
        } else { return nil, err }
    case IoTypeCodeFloat32:
        if f, err := r.ReadFloat32(); err == nil { 
            return Float32( f ), nil
        } else { return nil, err }
    case IoTypeCodeFloat64:
        if f, err := r.ReadFloat64(); err == nil { 
            return Float64( f ), nil
        } else { return nil, err }
    case IoTypeCodeBool:
        if b, err := r.ReadBool(); err == nil { 
            return Boolean( b ), nil
        } else { return nil, err }
    case IoTypeCodeEnum: return r.readEnum()
    }
    panic( libErrorf( "Not a scalar val type: 0x%02x", tc ) )
}

// Note: type code is already read
func ( r *BinReader ) readRegexRestriction() ( rr *RegexRestriction,
                                               err error ) {
    var src string
    if src, err = r.ReadUtf8(); err != nil { return }
    return CreateRegexRestriction( src )
}

func ( r *BinReader ) readRangeVal() ( Value, error ) {
    tc, err := r.ReadTypeCode()
    if err != nil { return nil, err }
    switch tc {
    case IoTypeCodeString, IoTypeCodeTimestamp, IoTypeCodeInt32, 
         IoTypeCodeInt64, IoTypeCodeUint32, IoTypeCodeUint64, IoTypeCodeFloat32,
         IoTypeCodeFloat64: 
        return r.ReadScalarValue( tc )
    case IoTypeCodeNull: 
        if _, err := r.ReadScalarValue( tc ); err != nil { return nil, err }
        return nil, nil
    }
    err = r.IoErrorf( "mingle: Unrecognized range value code: 0x%02x", tc )
    return nil, err
} 

// Note: type code is already read
func ( r *BinReader ) readRangeRestriction() ( rr *RangeRestriction,
                                               err error ) {

    var minClosed, maxClosed bool
    var min, max Value
    if minClosed, err = r.readBool(); err != nil { return }
    if min, err = r.readRangeVal(); err != nil { return }
    if max, err = r.readRangeVal(); err != nil { return }
    if maxClosed, err = r.readBool(); err != nil { return }
    return NewRangeRestriction( minClosed, min, max, maxClosed ), nil
}

func ( r *BinReader ) readRestriction() ( vr ValueRestriction, err error ) {
    var tc IoTypeCode 
    if tc, err = r.ReadTypeCode(); err != nil { return }
    switch tc {
    case IoTypeCodeNull: return nil, nil
    case IoTypeCodeRegexRestrict: return r.readRegexRestriction()
    case IoTypeCodeRangeRestrict: return r.readRangeRestriction()
    }
    err = fmt.Errorf( "mingle: Unrecognized restriction type code: 0x%02x", tc )
    return
}

func ( r *BinReader ) ReadAtomicTypeReference() ( at *AtomicTypeReference,
                                                  err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeAtomTyp ); err != nil { return }
    var nm *QualifiedTypeName
    var rx ValueRestriction
    if nm, err = r.ReadQualifiedTypeName(); err != nil { return }
    if rx, err = r.readRestriction(); err != nil { return }
    at = NewAtomicTypeReference( nm, rx )
    return
}

func ( r *BinReader ) ReadListTypeReference() ( lt *ListTypeReference,
                                               err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeListTyp ); err != nil { return }
    lt = &ListTypeReference{}
    if lt.ElementType, err = r.ReadTypeReference(); err != nil { return }
    if lt.AllowsEmpty, err = r.readBool(); err != nil { return }
    return
}

func ( r *BinReader ) ReadNullableTypeReference() ( nt *NullableTypeReference,
                                                    err error ) {
    if _, err = r.ExpectTypeCode( IoTypeCodeNullableTyp ); err != nil { return }
    var typ TypeReference
    if typ, err = r.ReadTypeReference(); err != nil { return }
    nt = MustNullableTypeReference( typ )
    return
}

func ( r *BinReader ) ReadPointerTypeReference() ( 
    pt *PointerTypeReference, err error ) {

    if _, err = r.ExpectTypeCode( IoTypeCodePointerTyp ); err != nil { return }
    var typ TypeReference
    if typ, err = r.ReadTypeReference(); err != nil { return }
    return NewPointerTypeReference( typ ), nil
}

func ( r *BinReader ) ReadTypeReference() ( typ TypeReference, err error ) {
    var tc IoTypeCode
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case IoTypeCodeAtomTyp: return r.ReadAtomicTypeReference()
    case IoTypeCodeListTyp: return r.ReadListTypeReference()
    case IoTypeCodeNullableTyp: return r.ReadNullableTypeReference()
    case IoTypeCodePointerTyp: return r.ReadPointerTypeReference()
    }
    err = r.IoErrorf( "Unrecognized type reference code: 0x%02x", tc )
    return
}

func TypeReferenceFromBytes( buf []byte ) ( TypeReference, error ) {
    return NewReaderBytes( buf ).ReadTypeReference()
}

func IdentifierFromBytes( buf []byte ) ( *Identifier, error ) {
    return NewReaderBytes( buf ).ReadIdentifier()
}

func QualifiedTypeNameFromBytes( buf []byte ) ( *QualifiedTypeName, error ) {
    return NewReaderBytes( buf ).ReadQualifiedTypeName()
}

func NamespaceFromBytes( buf []byte ) ( *Namespace, error ) {
    return NewReaderBytes( buf ).ReadNamespace()
}
