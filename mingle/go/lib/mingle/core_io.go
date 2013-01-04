package mingle

import (
    "fmt"
    "bytes"
    "io"
    "bitgirder/objpath"
    "time"
//    "log"
    bgio "bitgirder/io"
)

const (
    tcNull = uint8( 0x00 )
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
    tcTimestamp = uint8( 0x12 )
    tcBuffer = uint8( 0x13 )
    tcEnum = uint8( 0x14 )
    tcSymMap = uint8( 0x15 )
    tcField = uint8( 0x16 )
    tcStruct = uint8( 0x17 )
    tcList = uint8( 0x19 )
    tcEnd = uint8( 0x1a )
    tcIdPath = uint8( 0x1b )
    tcIdPathListNode = uint8( 0x1c )
)

type BinIoError struct { msg string }

func ( e *BinIoError ) Error() string { return e.msg }

type BinWriter struct { *bgio.BinWriter }

func AsWriter( w *bgio.BinWriter ) *BinWriter { return &BinWriter{ w } }

func NewWriter( w io.Writer ) *BinWriter { 
    return AsWriter( bgio.NewLeWriter( w ) )
}

func ( w *BinWriter ) WriteTypeCode( tc uint8 ) error {
    return w.WriteUint8( tc )
}

func ( w *BinWriter ) WriteNull() error { return w.WriteTypeCode( tcNull ) }

func ( w *BinWriter ) writeBool( b bool ) ( err error ) {
    return w.WriteValue( Boolean( b ) )
}

func ( w *BinWriter ) WriteIdentifier( id *Identifier ) ( err error ) {
    if err = w.WriteTypeCode( tcId ); err != nil { return }
    if err = w.WriteUint8( uint8( len( id.parts ) ) ); err != nil { return }
    for _, part := range id.parts {
        if err = w.WriteBuffer32( part ); err != nil { return }
    }
    return
}

func ( w *BinWriter ) writeIds( ids []*Identifier ) ( err error ) {
    if err = w.WriteUint8( uint8( len( ids ) ) ); err != nil { return }
    for _, id := range ids {
        if err = w.WriteIdentifier( id ); err != nil { return }
    }
    return
}

type pathWriter struct { w *BinWriter }

// Write the tcId even though WriteIdentifier does so that id path reads can
// unconditionally read a type code as they go
func ( pw pathWriter ) Descend( elt interface{} ) ( err error ) {
    if err = pw.w.WriteTypeCode( tcId ); err != nil { return }
    return pw.w.WriteIdentifier( elt.( *Identifier ) )
}

func ( pw pathWriter ) List( idx int ) ( err error ) {
    if err = pw.w.WriteTypeCode( tcIdPathListNode ); err != nil { return }
    return pw.w.WriteInt32( int32( idx ) )
}

func ( w *BinWriter ) WriteIdPath( p objpath.PathNode ) ( err error ) {
    if err = w.WriteTypeCode( tcIdPath ); err != nil { return }
    if err = objpath.Visit( p, pathWriter{ w } ); err != nil { return }
    return w.WriteTypeCode( tcEnd )
}

func ( w *BinWriter ) WriteNamespace( ns *Namespace ) ( err error ) {
    if err = w.WriteTypeCode( tcNs ); err != nil { return }
    if err = w.writeIds( ns.Parts ); err != nil { return }
    return w.WriteIdentifier( ns.Version )
}

func ( w *BinWriter ) WriteDeclaredTypeName( 
    n *DeclaredTypeName ) ( err error ) {
    if err = w.WriteTypeCode( tcDeclNm ); err != nil { return }
    return w.WriteBuffer32( n.nm )
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
    return w.WriteUtf8( rr.src )
}

func ( w *BinWriter ) writeEnum( en *Enum ) ( err error ) {
    if err = w.WriteTypeCode( tcEnum ); err != nil { return }
    if err = w.WriteTypeReference( en.Type ); err != nil { return }
    if err = w.WriteIdentifier( en.Value ); err != nil { return }
    return
}

type writeReactor struct { *BinWriter }

func ( w writeReactor ) startStruct( typ TypeReference ) error {
    if err := w.WriteTypeCode( tcStruct ); err != nil { return err }
    if err := w.WriteInt32( int32( -1 ) ); err != nil { return err }
    return w.WriteTypeReference( typ )
}

func ( w writeReactor ) startField( fld *Identifier ) error {
    if err := w.WriteTypeCode( tcField ); err != nil { return err }
    return w.WriteIdentifier( fld )
}

func ( w writeReactor ) startList() error { 
    if err := w.WriteTypeCode( tcList ); err != nil { return err }
    return w.WriteInt32( -1 )
}

func ( w writeReactor ) startMap() error { return w.WriteTypeCode( tcSymMap ) }

func ( w writeReactor ) value( val Value ) error {
    switch v := val.( type ) {
    case nil: return w.WriteNull()
    case *Null: return w.WriteNull()
    case Boolean: 
        if err := w.WriteTypeCode( tcBool ); err != nil { return err }
        return w.WriteBool( bool( v ) )
    case Buffer:
        if err := w.WriteTypeCode( tcBuffer ); err != nil { return err }
        return w.WriteBuffer32( []byte( v ) )
    case String:
        if err := w.WriteTypeCode( tcString ); err != nil { return err }
        return w.WriteUtf8( string( v ) )
    case Int32:
        if err := w.WriteTypeCode( tcInt32 ); err != nil { return err }
        return w.WriteInt32( int32( v ) )
    case Int64:
        if err := w.WriteTypeCode( tcInt64 ); err != nil { return err }
        return w.WriteInt64( int64( v ) )
    case Uint32:
        if err := w.WriteTypeCode( tcUint32 ); err != nil { return err }
        return w.WriteUint32( uint32( v ) )
    case Uint64:
        if err := w.WriteTypeCode( tcUint64 ); err != nil { return err }
        return w.WriteUint64( uint64( v ) )
    case Float32:
        if err := w.WriteTypeCode( tcFloat32 ); err != nil { return err }
        return w.WriteFloat32( float32( v ) )
    case Float64:
        if err := w.WriteTypeCode( tcFloat64 ); err != nil { return err }
        return w.WriteFloat64( float64( v ) )
    case Timestamp:
        if err := w.WriteTypeCode( tcTimestamp ); err != nil { return err }
        if err := w.WriteInt64( time.Time( v ).Unix() ); err != nil { 
            return err
        }
        return w.WriteInt32( int32( time.Time( v ).Nanosecond() ) )
    case *Enum: return w.writeEnum( v )
    }
    panic( libErrorf( "%T: Unhandled value: %T", w, val ) )
}

func ( w writeReactor ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case ValueEvent: return w.value( v.Val )
    case MapStartEvent: return w.startMap()
    case StructStartEvent: return w.startStruct( v.Type )
    case ListStartEvent: return w.startList()
    case FieldStartEvent: return w.startField( v.Field )
    case EndEvent: return w.WriteTypeCode( tcEnd )
    }
    panic( libErrorf( "Unhandled event type: %T", ev ) )
}

func ( w *BinWriter ) AsReactor() ReactorEventProcessor { 
    return writeReactor{ w } 
}

func ( w *BinWriter ) WriteValue( val Value ) ( err error ) {
    return VisitValue( val, w.AsReactor() )
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
    case nil: return w.WriteNull()
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

func ( r *BinReader ) offset() int64 {
    return r.ot.off
}

func ( r *BinReader ) ioErrorf( tmpl string, args ...interface{} ) *BinIoError {
    str := &bytes.Buffer{}
    fmt.Fprintf( str, "[offset %d]: ", r.offset() - 1 )
    fmt.Fprintf( str, tmpl, args... )
    return &BinIoError{ str.String() }
}

func ( r *BinReader ) ReadTypeCode() ( res uint8, err error ) {
//    if r.tcSaved < 0 { return r.ReadUint8() }
//    res, err, r.tcSaved = uint8( r.tcSaved ), nil, -1
//    return res, err
    return r.ReadUint8()
}

// State of reader is undefined after a call to this method that returns a
// non-nil error
func ( r *BinReader ) PeekTypeCode() ( uint8, error ) {
    res, err := r.ReadTypeCode()
    if err2 := r.ot.UnreadByte(); err2 != nil { panic( err2 ) }
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
    return r.ReadBool()
}

func ( r *BinReader ) ReadIdentifier() ( id *Identifier, err error ) {
    if _, err = r.ExpectTypeCode( tcId ); err != nil { return }
    var sz uint8
    if sz, err = r.ReadUint8(); err != nil { return }
    id = &Identifier{ make( []idPart, sz ) }
    for i := uint8( 0 ); i < sz; i++ {
        if id.parts[ i ], err = r.ReadBuffer32(); err != nil { return }
    }
    return    
}

func ( r *BinReader ) readIds() ( ids []*Identifier, err error ) {
    var sz uint8
    if sz, err = r.ReadUint8(); err != nil { return }
    ids = make( []*Identifier, sz )
    for i := uint8( 0 ); i < sz; i++ {
        if ids[ i ], err = r.ReadIdentifier(); err != nil { return }
    }
    return
}

func ( r *BinReader ) readIdPathNext( 
    p objpath.PathNode ) ( objpath.PathNode, bool, error ) {
    tc, err := r.ReadTypeCode()
    if err != nil { return nil, false, err }
    switch tc {
    case tcId:
        if id, err := r.ReadIdentifier(); err == nil { 
            if p == nil { 
                return objpath.RootedAt( id ), false, nil
            } else { return p.Descend( id ), false, nil }
        } else { return nil, false, err }
    case tcIdPathListNode:
        if i, err := r.ReadInt32(); err == nil {
            var l *objpath.ListNode
            if p == nil { 
                l = objpath.RootedAtList() 
            } else { l = p.StartList() }
            for ; i > 0; i-- { l = l.Next() }
            return l, false, nil
        } else { return nil, false, err }
    case tcEnd: return p, true, nil
    }
    return nil, false, r.ioErrorf( "Unrecognized id path code: 0x%02x", tc )
}

func ( r *BinReader ) ReadIdPath() ( p objpath.PathNode, err error ) {
    if _, err = r.ExpectTypeCode( tcIdPath ); err != nil { return }
    for done := false; ! done; {
        if p, done, err = r.readIdPathNext( p ); err != nil { return }
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
    if buf, err = r.ReadBuffer32(); err != nil { return }
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
    if src, err = r.ReadUtf8(); err != nil { return }
    return NewRegexRestriction( src )
}

func ( r *BinReader ) readEnum() ( en *Enum, err error ) {
    en = &Enum{}
    if en.Type, err = r.ReadTypeReference(); err != nil { return }
    if en.Value, err = r.ReadIdentifier(); err != nil { return }
    return
}

func ( r *BinReader ) readScalarValue( 
    tc uint8, rep ReactorEventProcessor ) ( err error ) {
    var val Value
    switch tc {
    case tcNull: val = NullVal
    case tcString: 
        var s string
        if s, err = r.ReadUtf8(); err == nil { val = String( s ) }
    case tcBuffer:
        var buf []byte
        if buf, err = r.ReadBuffer32(); err == nil { val = Buffer( buf ) }
    case tcTimestamp:
        var ux int64
        var ns int32
        if ux, err = r.ReadInt64(); err == nil { 
            if ns, err = r.ReadInt32(); err == nil { 
                val = Timestamp( time.Unix( ux, int64( ns ) ) )
            } 
        }
    case tcInt32:
        var i int32
        if i, err = r.ReadInt32(); err == nil { val = Int32( i ) }
    case tcInt64:
        var i int64
        if i, err = r.ReadInt64(); err == nil { val = Int64( i ) }
    case tcUint32:
        var i uint32
        if i, err = r.ReadUint32(); err == nil { val = Uint32( i ) }
    case tcUint64:
        var i uint64
        if i, err = r.ReadUint64(); err == nil { val = Uint64( i ) }
    case tcFloat32:
        var f float32
        if f, err = r.ReadFloat32(); err == nil { val = Float32( f ) }
    case tcFloat64:
        var f float64
        if f, err = r.ReadFloat64(); err == nil { val = Float64( f ) }
    case tcBool:
        var b bool
        if b, err = r.ReadBool(); err == nil { val = Boolean( b ) }
    case tcEnum: val, err = r.readEnum()
    default: panic( libErrorf( "Not a scalar val type: 0x%02x", tc ) )
    }
    if err == nil { err = rep.ProcessEvent( ValueEvent{ val } ) }
    return 
}

func ( r *BinReader ) readMapFields( rep ReactorEventProcessor ) error {
    for {
        tc, err := r.ReadTypeCode()
        if err != nil { return err }
        switch tc {
        case tcEnd: return rep.ProcessEvent( EvEnd )
        case tcField:
            id, err := r.ReadIdentifier()
            if err == nil { err = rep.ProcessEvent( FieldStartEvent{ id } ) }
            if err != nil { return err }
            if err := r.implReadValue( rep ); err != nil { return err }
        default: return r.ioErrorf( "Unexpected map pair code: 0x%02x", tc )
        }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( r *BinReader ) readSymbolMap( rep ReactorEventProcessor ) error {
    if err := rep.ProcessEvent( EvMapStart ); err != nil { return err }
    return r.readMapFields( rep )
}

func ( r *BinReader ) readStruct( rep ReactorEventProcessor ) error {
    if _, err := r.ReadInt32(); err != nil { return err }
    if typ, err := r.ReadTypeReference(); err == nil {
        if err = rep.ProcessEvent( StructStartEvent{ typ } ); err != nil { 
            return err 
        }
    } else { return err }
    return r.readMapFields( rep )
}

func ( r *BinReader ) readList( rep ReactorEventProcessor ) error {
    if _, err := r.ReadInt32(); err != nil { return err } // skip size
    if err := rep.ProcessEvent( EvListStart ); err != nil { return err }
    for {
        tc, err := r.PeekTypeCode()
        if err != nil { return err }
        if tc == tcEnd {
            if _, err = r.ReadTypeCode(); err != nil { return err }
            return rep.ProcessEvent( EvEnd )
        } else { 
            if err = r.implReadValue( rep ); err != nil { return err } 
        }
    }
    panic( libErrorf( "Unreachable" ) )
}

func ( r *BinReader ) implReadValue( rep ReactorEventProcessor ) error {
    tc, err := r.ReadTypeCode()
    if err != nil { return err }
    switch tc {
    case tcNull, tcString, tcBuffer, tcTimestamp, tcInt32, tcInt64, tcUint32,
         tcUint64, tcFloat32, tcFloat64, tcBool, tcEnum:
        return r.readScalarValue( tc, rep )
    case tcSymMap: return r.readSymbolMap( rep )
    case tcStruct: return r.readStruct( rep )
    case tcList: return r.readList( rep )
    default: return r.ioErrorf( "Unrecognized value code: 0x%02x", tc )
    }
    panic( libErrorf( "Unreachable" ) )
}

func ( r *BinReader ) ReadReactorValue( rep ReactorEventProcessor ) error {
    return r.implReadValue( rep )
}

func ( r *BinReader ) ReadValue() ( Value, error ) {
    vb := NewValueBuilder()
    pip := InitReactorPipeline( vb )
//    vb := NewValueBuilder()
//    vb.SetTopType( ReactorTopTypeValue )
//    err := r.ReadReactorValue( vb )
//    if err != nil { return nil, err }
//    return vb.GetValue(), nil
    err := r.ReadReactorValue( pip )
    if err != nil { return nil, err }
    return vb.GetValue(), nil
}

func ( r *BinReader ) readRangeVal() ( val Value, err error ) {
    var tc uint8
    if tc, err = r.PeekTypeCode(); err != nil { return }
    switch tc {
    case tcString, tcTimestamp, tcInt32, tcInt64, tcUint32, tcUint64,
         tcFloat32, tcFloat64: 
        return r.ReadValue()
    case tcNull: 
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
    case tcNull: return nil, nil
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
    err = r.ioErrorf( "Unrecognized type reference code: 0x%02x", tc )
    return
}
