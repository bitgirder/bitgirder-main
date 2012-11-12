package mingle

import (
    "fmt"
    "io"
//    "log"
    bgio "bitgirder/io"
)

const (
    tcNil = uint8( 0x00 )
    tcId = uint8( 0x01 )
    tcNs = uint8( 0x02 )
    tcDeclNm = uint8( 0x03 )
    tcQn = uint8( 0x04 )
    tcAtomTyp = uint8( 0x05 )
    tcListTyp = uint8( 0x06 )
    tcNullableTyp = uint8( 0x07 )
    tcRegexRestrict = uint8( 0x08 )
    tcRangeRestrict = uint8( 0x09 )
    tcBool = uint8( 0x0a )
    tcString = uint8( 0x0b )
    tcInt32 = uint8( 0x0c )
    tcInt64 = uint8( 0x0d )
    tcUint32 = uint8( 0x0e )
    tcUint64 = uint8( 0x0f )
    tcFloat32 = uint8( 0x10 )
    tcFloat64 = uint8( 0x11 )
    tcTimeRfc3339 = uint8( 0x12 )
    tcBuffer = uint8( 0x13 )
    tcEnum = uint8( 0x14 )
    tcSymMap = uint8( 0x15 )
    tcMapPair = uint8( 0x16 )
    tcStruct = uint8( 0x17 )
    tcList = uint8( 0x19 )
    tcEnd = uint8( 0x1a )
)

type BinWriter struct { w *bgio.BinWriter }

func AsWriter( w *bgio.BinWriter ) *BinWriter { return &BinWriter{ w } }

func NewWriter( w io.Writer ) *BinWriter { 
    return AsWriter( bgio.NewLeWriter( w ) )
}

func ( w *BinWriter ) WriteTypeCode( tc uint8 ) error {
    return w.w.WriteUint8( tc )
}

func ( w *BinWriter ) WriteNil() error { return w.WriteTypeCode( tcNil ) }

func ( w *BinWriter ) writeBool( b bool ) ( err error ) {
    return w.WriteValue( Boolean( b ) )
}

func ( w *BinWriter ) WriteIdentifier( id *Identifier ) ( err error ) {
    if err = w.WriteTypeCode( tcId ); err != nil { return }
    if err = w.w.WriteUint8( uint8( len( id.parts ) ) ); err != nil { return }
    for _, part := range id.parts {
        if err = w.w.WriteBuffer32( part ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeIds( ids []*Identifier ) ( err error ) {
    if err = w.w.WriteUint8( uint8( len( ids ) ) ); err != nil { return }
    for _, id := range ids {
        if err = w.WriteIdentifier( id ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) WriteNamespace( ns *Namespace ) ( err error ) {
    if err = w.WriteTypeCode( tcNs ); err != nil { return }
    if err = w.writeIds( ns.Parts ); err != nil { return }
    return w.WriteIdentifier( ns.Version )
}

func ( w *BinWriter ) WriteDeclaredTypeName( 
    n *DeclaredTypeName ) ( err error ) {
    if err = w.WriteTypeCode( tcDeclNm ); err != nil { return }
    return w.w.WriteBuffer32( n.nm )
}

func ( w *BinWriter ) WriteQualifiedTypeName( 
    qn *QualifiedTypeName ) ( err error ) {
    if err = w.WriteTypeCode( tcQn ); err != nil { return }
    if err = w.WriteNamespace( qn.Namespace ); err != nil { return }
    return w.WriteDeclaredTypeName( qn.Name )
}

func ( w *BinWriter ) WriteTypeName( nm TypeName ) error {
    switch v := nm.( type ) {
    case *DeclaredTypeName: return w.WriteDeclaredTypeName( v )
    case *QualifiedTypeName: return w.WriteQualifiedTypeName( v )
    }
    panic( fmt.Errorf( "%T: Unhandled type name: %T", w, nm ) )
}

func ( w *BinWriter ) writeRegexRestriction( 
    rr *RegexRestriction ) ( err error ) {
    if err = w.WriteTypeCode( tcRegexRestrict ); err != nil { return }
    return w.w.WriteUtf8( rr.src )
}

func ( w *BinWriter ) writeEnum( en *Enum ) ( err error ) {
    if err = w.WriteTypeCode( tcEnum ); err != nil { return }
    if err = w.WriteTypeReference( en.Type ); err != nil { return }
    if err = w.WriteIdentifier( en.Value ); err != nil { return }
    return
}

func ( w *BinWriter ) writeSymbolMap( m *SymbolMap ) ( err error ) {
    if err = w.WriteTypeCode( tcSymMap ); err != nil { return }
    err = m.EachPairError( func( k *Identifier, v Value ) error {
        if err2 := w.WriteTypeCode( tcMapPair ); err2 != nil { return err2 }
        if err2 := w.WriteIdentifier( k ); err2 != nil { return err2 }
        return w.WriteValue( v )
    })
    if err != nil { return }
    return w.WriteTypeCode( tcEnd )
}

func ( w *BinWriter ) writeStruct( ms *Struct ) ( err error ) {
    if err = w.WriteTypeCode( tcStruct ); err != nil { return }
    if err = w.WriteTypeReference( ms.Type ); err != nil { return }
    if err = w.WriteValue( ms.Fields ); err != nil { return }
    return
}

func ( w *BinWriter ) writeList( l *List ) ( err error ) {
    if err = w.WriteTypeCode( tcList ); err != nil { return }
    for _, val := range l.Values() {
        if err = w.WriteValue( val ); err != nil { return }
    }
    return w.WriteTypeCode( tcEnd )
}

func ( w *BinWriter ) WriteValue( val Value ) ( err error ) {
    switch v := val.( type ) {
    case nil: return w.WriteNil()
    case *Null: return w.WriteNil()
    case Boolean: 
        if err = w.WriteTypeCode( tcBool ); err != nil { return }
        return w.w.WriteBool( bool( v ) )
    case Buffer:
        if err = w.WriteTypeCode( tcBuffer ); err != nil { return }
        return w.w.WriteBuffer32( []byte( v ) )
    case String:
        if err = w.WriteTypeCode( tcString ); err != nil { return }
        return w.w.WriteUtf8( string( v ) )
    case Int32:
        if err = w.WriteTypeCode( tcInt32 ); err != nil { return }
        return w.w.WriteInt32( int32( v ) )
    case Int64:
        if err = w.WriteTypeCode( tcInt64 ); err != nil { return }
        return w.w.WriteInt64( int64( v ) )
    case Uint32:
        if err = w.WriteTypeCode( tcUint32 ); err != nil { return }
        return w.w.WriteUint32( uint32( v ) )
    case Uint64:
        if err = w.WriteTypeCode( tcUint64 ); err != nil { return }
        return w.w.WriteUint64( uint64( v ) )
    case Float32:
        if err = w.WriteTypeCode( tcFloat32 ); err != nil { return }
        return w.w.WriteFloat32( float32( v ) )
    case Float64:
        if err = w.WriteTypeCode( tcFloat64 ); err != nil { return }
        return w.w.WriteFloat64( float64( v ) )
    case Timestamp:
        if err = w.WriteTypeCode( tcTimeRfc3339 ); err != nil { return }
        return w.w.WriteUtf8( v.Rfc3339Nano() )
    case *Enum: return w.writeEnum( v )
    case *SymbolMap: return w.writeSymbolMap( v )
    case *Struct: return w.writeStruct( v )
    case *List: return w.writeList( v )
    }
    panic( fmt.Errorf( "%T: Unhandled value: %T", w, val ) )
}

func ( w *BinWriter ) writeRangeValue( val Value ) error {
    switch val.( type ) {
    case nil, Null, String, Int32, Int64, Uint32, Uint64, Float32, Float64,
         Timestamp:
        return w.WriteValue( val )
    }
    panic( fmt.Errorf( "%T: Unhandled range val: %T", w, val ) )
}

func ( w *BinWriter ) writeRangeRestriction( 
    rr *RangeRestriction ) ( err error ) {
    if err = w.WriteTypeCode( tcRangeRestrict ); err != nil { return }
    if err = w.writeBool( rr.MinClosed ); err != nil { return }
    if err = w.writeRangeValue( rr.Min ); err != nil { return }
    if err = w.writeRangeValue( rr.Max ); err != nil { return }
    return w.writeBool( rr.MaxClosed )
}

func ( w *BinWriter ) WriteAtomicTypeReference( 
    at *AtomicTypeReference ) ( err error ) {
    if err = w.WriteTypeCode( tcAtomTyp ); err != nil { return }
    if err = w.WriteTypeName( at.Name ); err != nil { return }
    switch r := at.Restriction.( type ) {
    case nil: return w.WriteNil()
    case *RegexRestriction: return w.writeRegexRestriction( r )
    case *RangeRestriction: return w.writeRangeRestriction( r )
    default: panic( fmt.Errorf( "%T: Unhandled restriction: %T", w, r ) )
    }
    return
}

func ( w *BinWriter ) WriteListTypeReference( 
    lt *ListTypeReference ) ( err error ) {
    if err = w.WriteTypeCode( tcListTyp ); err != nil { return }
    if err = w.WriteTypeReference( lt.ElementType ); err != nil { return }
    return w.writeBool( lt.AllowsEmpty )
}

func ( w *BinWriter ) WriteNullableTypeReference( 
    nt *NullableTypeReference ) ( err error ) {
    if err = w.WriteTypeCode( tcNullableTyp ); err != nil { return }
    return w.WriteTypeReference( nt.Type )
}

func ( w *BinWriter ) WriteTypeReference( typ TypeReference ) error {
    switch v := typ.( type ) {
    case *AtomicTypeReference: return w.WriteAtomicTypeReference( v )
    case *ListTypeReference: return w.WriteListTypeReference( v )
    case *NullableTypeReference: return w.WriteNullableTypeReference( v )
    }
    panic( fmt.Errorf( "%T: Unhandled type reference: %T", w, typ ) )
}

type BinReader struct {
    r *bgio.BinReader
    tcSaved int16
}

func AsReader( r *bgio.BinReader ) *BinReader { 
    return &BinReader{ r: r, tcSaved: -1 } 
}

func NewReader( r io.Reader ) *BinReader {
    return AsReader( bgio.NewLeReader( r ) )
}

func ( r *BinReader ) ReadTypeCode() ( res uint8, err error ) {
    if r.tcSaved < 0 { return r.r.ReadUint8() }
    res, err, r.tcSaved = uint8( r.tcSaved ), nil, -1
    return res, err
}

// State of reader is undefined after a call to this method that returns a
// non-nil error
func ( r *BinReader ) PeekTypeCode() ( uint8, error ) {
    res, err := r.ReadTypeCode()
    r.tcSaved = int16( res )
    return res, err
}

func ( r *BinReader ) ExpectTypeCode( expct uint8 ) ( res uint8, err error ) {
    if res, err = r.ReadTypeCode(); err == nil {
        if res != expct { 
            tmpl := "Expected type code 0x%02x but got 0x%02x"
            err = fmt.Errorf( tmpl, expct, res )
        }
    }
    return
}

func ( r *BinReader ) readBool() ( bool, error ) {
    if _, err := r.ExpectTypeCode( tcBool ); err != nil { return false, err }
    return r.r.ReadBool()
}

func ( r *BinReader ) ReadIdentifier() ( id *Identifier, err error ) {
    if _, err = r.ExpectTypeCode( tcId ); err != nil { return }
    var sz uint8
    if sz, err = r.r.ReadUint8(); err != nil { return }
    id = &Identifier{ make( []idPart, sz ) }
    for i := uint8( 0 ); i < sz; i++ {
        if id.parts[ i ], err = r.r.ReadBuffer32(); err != nil { return }
    }
    return    
}

func ( r *BinReader ) readIds() ( ids []*Identifier, err error ) {
    var sz uint8
    if sz, err = r.r.ReadUint8(); err != nil { return }
    ids = make( []*Identifier, sz )
    for i := uint8( 0 ); i < sz; i++ {
        if ids[ i ], err = r.ReadIdentifier(); err != nil { return }
    }
    return
}

func ( r *BinReader ) ReadNamespace() ( ns *Namespace, err error ) {
    if _, err = r.ExpectTypeCode( tcNs ); err != nil { return }
    ns = &Namespace{}
    if ns.Parts, err = r.readIds(); err != nil { return }
    if ns.Version, err = r.ReadIdentifier(); err != nil { return }
    return
}

func ( r *BinReader ) ReadDeclaredTypeName() ( nm *DeclaredTypeName,    
                                               err error ) {
    if _, err = r.ExpectTypeCode( tcDeclNm ); err != nil { return }
    var buf []byte
    if buf, err = r.r.ReadBuffer32(); err != nil { return }
    nm = &DeclaredTypeName{ buf }
    return
}

func ( r *BinReader ) ReadQualifiedTypeName() ( qn *QualifiedTypeName,
                                                err error ) {
    if _, err = r.ExpectTypeCode( tcQn ); err != nil { return }
    qn = &QualifiedTypeName{}
    if qn.Namespace, err = r.ReadNamespace(); err != nil { return }
    if qn.Name, err = r.ReadDeclaredTypeName(); err != nil { return }
    return
}

func ( r *BinReader ) ReadTypeName() ( nm TypeName, err error ) {
    var tc uint8
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case tcDeclNm: return r.ReadDeclaredTypeName()
    case tcQn: return r.ReadQualifiedTypeName()
    }
    return nil, fmt.Errorf( "mingle: Unrecognized type name code: 0x%02x", tc )
}

// Note: type code is already read
func ( r *BinReader ) readRegexRestriction() ( rr *RegexRestriction,
                                               err error ) {
    var src string
    if src, err = r.r.ReadUtf8(); err != nil { return }
    return NewRegexRestriction( src )
}

func ( r *BinReader ) readEnum() ( en *Enum, err error ) {
    en = &Enum{}
    if en.Type, err = r.ReadTypeReference(); err != nil { return }
    if en.Value, err = r.ReadIdentifier(); err != nil { return }
    return
}

func ( r *BinReader ) readSymbolMap() ( m *SymbolMap, err error ) {
    flds := make( []fieldEntry, 0, 8 )
    for m == nil && err == nil {
        var tc uint8
        if tc, err = r.ReadTypeCode(); err != nil { return }
        switch tc {
        case tcEnd: m, err = makeSymbolMap( flds ) 
        case tcMapPair:
            e := fieldEntry{}
            if e.id, err = r.ReadIdentifier(); err != nil { return }
            if e.val, err = r.ReadValue(); err != nil { return }
            flds = append( flds, e )
        default: err = fmt.Errorf( "Unexpected map pair code: 0x%02x", tc )
        }
    }
    return
}

func ( r *BinReader ) readStruct() ( val *Struct, err error ) {
    val = &Struct{}
    if val.Type, err = r.ReadTypeReference(); err != nil { return }
    var flds Value
    if flds, err = r.expectValue( tcSymMap ); err != nil { return }
    val.Fields = flds.( *SymbolMap )
    return
}

func ( r *BinReader ) readList() ( l *List, err error ) {
    vals := make( []Value, 0, 4 )
    for l == nil {
        var tc uint8
        if tc, err = r.PeekTypeCode(); err != nil { return }
        if tc == tcEnd { 
            if _, err = r.ReadTypeCode(); err != nil { return } // consume it
            l = NewList( vals ) 
        } else {
            var val Value
            if val, err = r.ReadValue(); err != nil { return }
            vals = append( vals, val )
        }
    }
    return
}

func ( r *BinReader ) implReadValue( tc uint8 ) ( val Value, err error ) {
    switch tc {
    case tcNil: val = NullVal
    case tcString:
        var s string
        if s, err = r.r.ReadUtf8(); err == nil { val = String( s ) }
    case tcBuffer:
        var buf []byte
        if buf, err = r.r.ReadBuffer32(); err == nil { val = Buffer( buf ) }
    case tcTimeRfc3339:
        var s string
        if s, err = r.r.ReadUtf8(); err == nil { 
            val, err = ParseTimestamp( s ) 
        }
    case tcInt32:
        var i int32
        if i, err = r.r.ReadInt32(); err == nil { val = Int32( i ) }
    case tcInt64:
        var i int64
        if i, err = r.r.ReadInt64(); err == nil { val = Int64( i ) }
    case tcUint32:
        var i uint32
        if i, err = r.r.ReadUint32(); err == nil { val = Uint32( i ) }
    case tcUint64:
        var i uint64
        if i, err = r.r.ReadUint64(); err == nil { val = Uint64( i ) }
    case tcFloat32:
        var f float32
        if f, err = r.r.ReadFloat32(); err == nil { val = Float32( f ) }
    case tcFloat64:
        var f float64
        if f, err = r.r.ReadFloat64(); err == nil { val = Float64( f ) }
    case tcBool:
        var b bool
        if b, err = r.r.ReadBool(); err == nil { val = Boolean( b ) }
    case tcEnum: val, err = r.readEnum()
    case tcSymMap: val, err = r.readSymbolMap()
    case tcStruct: val, err = r.readStruct()
    case tcList: val, err = r.readList()
    default:
        err = fmt.Errorf( "mingle: Unrecognized value code: 0x%02x", tc )
    }
    return
}

func ( r *BinReader ) ReadValue() ( val Value, err error ) {
    var tc uint8
    if tc, err = r.ReadTypeCode(); err != nil { return }
    return r.implReadValue( tc )
}

func ( r *BinReader ) expectValue( expct uint8 ) ( Value, error ) {
    tc, err := r.ReadTypeCode()
    if err != nil { return nil, err }
    if tc != expct {
        tmpl := "mingle: Expected type 0x%02x but found 0x%02x"
        return nil, fmt.Errorf( tmpl, expct, tc )
    }
    return r.implReadValue( tc )
}

func ( r *BinReader ) readRangeVal() ( val Value, err error ) {
    var tc uint8
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case tcString, tcTimeRfc3339, tcInt32, tcInt64, tcUint32, tcUint64,
         tcFloat32, tcFloat64: 
        return r.ReadValue()
    case tcNil: 
        _, err = r.ReadValue()
        return
    }
    err = fmt.Errorf( "mingle: Unrecognized range value code: 0x%02x", tc )
    return
} 

// Note: type code is already read
func ( r *BinReader ) readRangeRestriction() ( rr *RangeRestriction,
                                               err error ) {
    rr = &RangeRestriction{}
    if rr.MinClosed, err = r.readBool(); err != nil { return }
    if rr.Min, err = r.readRangeVal(); err != nil { return }
    if rr.Max, err = r.readRangeVal(); err != nil { return }
    if rr.MaxClosed, err = r.readBool(); err != nil { return }
    return
}

func ( r *BinReader ) readRestriction() ( vr ValueRestriction, err error ) {
    var tc uint8
    if tc, err = r.ReadTypeCode(); err != nil { return }
    switch tc {
    case tcNil: return nil, nil
    case tcRegexRestrict: return r.readRegexRestriction()
    case tcRangeRestrict: return r.readRangeRestriction()
    }
    err = fmt.Errorf( "mingle: Unrecognized restriction type code: 0x%02x", tc )
    return
}

func ( r *BinReader ) ReadAtomicTypeReference() ( at *AtomicTypeReference,
                                                  err error ) {
    if _, err = r.ExpectTypeCode( tcAtomTyp ); err != nil { return }
    at = &AtomicTypeReference{}
    if at.Name, err = r.ReadTypeName(); err != nil { return }
    if at.Restriction, err = r.readRestriction(); err != nil { return }
    return
}

func ( r *BinReader ) ReadListTypeReference() ( lt *ListTypeReference,
                                               err error ) {
    if _, err = r.ExpectTypeCode( tcListTyp ); err != nil { return }
    lt = &ListTypeReference{}
    if lt.ElementType, err = r.ReadTypeReference(); err != nil { return }
    if lt.AllowsEmpty, err = r.readBool(); err != nil { return }
    return
}

func ( r *BinReader ) ReadNullableTypeReference() ( nt *NullableTypeReference,
                                                    err error ) {
    if _, err = r.ExpectTypeCode( tcNullableTyp ); err != nil { return }
    nt = &NullableTypeReference{}
    if nt.Type, err = r.ReadTypeReference(); err != nil { return }
    return
}

func ( r *BinReader ) ReadTypeReference() ( typ TypeReference, err error ) {
    var tc uint8
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case tcAtomTyp: return r.ReadAtomicTypeReference()
    case tcListTyp: return r.ReadListTypeReference()
    case tcNullableTyp: return r.ReadNullableTypeReference()
    }
    err = fmt.Errorf( "mingle: Unrecognized type reference code: 0x%02x", tc )
    return
}
