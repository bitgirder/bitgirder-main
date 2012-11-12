package codec

import (
    "io"
    "fmt"
    "bytes"
    mg "mingle"
//    "log"
    "container/list"
)

// Misc design note: Reactor is meant to be composable, and the design should
// lend itself to allow reactor chaining to accomplish specific things. For
// instance, a binary or json codec may not itself track and generate errors
// upon being fed a symbol map with duplicate keys, since doing complicates its
// implementation and incurs memory/cpu cost. However, when callers want to
// strictly check this, they might wrap the reactor in another one that tracks
// field names and generates an error on a duplicate, but otherwise passes calls
// down. Other chaining examples might be reactors which check the call sequence
// for validity with some known schema or data definition.

type Reactor interface {
    Value( v mg.Value ) error
    StartStruct( typ mg.TypeReference ) error
    StartMap() error
    StartList() error
    StartField( fld *mg.Identifier ) error
    End() error
}

func getTopLevelStructStartExpectError( valName string ) error {
    return Errorf( "Expected top-level struct start but got %s", valName )
}

type reactorImplMap struct {
    pending *mg.Identifier
}

func newReactorImplMap() *reactorImplMap { return &reactorImplMap{} }

func ( m *reactorImplMap ) startField( fld *mg.Identifier ) error {
    if m.pending == nil {
        m.pending = fld
        return nil
    }
    tmpl := "Saw start of field '%s' while expecting a value for '%s'"
    return Errorf( tmpl, fld, m.pending )
}

// Clears m.pending on nil return val
func ( m *reactorImplMap ) checkValue( valName string ) error {
    if m.pending == nil {
        tmpl := "Expected field name or end of fields but got %s"
        return Errorf( tmpl, valName )
    }
    m.pending = nil
    return nil 
}

func ( m *reactorImplMap ) end() error {
    if m.pending == nil { return nil }
    return Errorf( "Saw end while expecting value for field '%s'", m.pending )
}

type ReactorImpl struct {
    stack *list.List
    done bool
}

func NewReactorImpl() *ReactorImpl {
    return &ReactorImpl{ stack: &list.List{} }
}

func ( ri *ReactorImpl ) checkActive( call string ) error {
    if ri.done { return Errorf( "%s() called, but struct is built", call ) }
    return nil
}

// ri.stack known to be non-empty when this returns without error
func ( ri *ReactorImpl ) checkValue( valName string ) error {
    if ri.stack.Len() == 0 {
        return getTopLevelStructStartExpectError( valName )
    }
    elt := ri.stack.Front().Value
    if m, ok := elt.( *reactorImplMap ); ok { return m.checkValue( valName ) }
    return nil
}

func ( ri *ReactorImpl ) push( elt interface{} ) { ri.stack.PushFront( elt ) }

func ( ri *ReactorImpl ) StartStruct() error {
    if err := ri.checkActive( "StartStruct" ); err != nil { return err }
    // skip check if we're pushing the top level struct
    if ri.stack.Len() > 0 {
        if err := ri.checkValue( "struct start" ); err != nil { return err }
    }
    ri.push( newReactorImplMap() )
    return nil
}

func ( ri *ReactorImpl ) StartMap() error {
    if err := ri.checkActive( "StartMap" ); err != nil { return err }
    if err := ri.checkValue( "map start" ); err != nil { return err }
    ri.push( newReactorImplMap() )
    return nil
}

func ( ri *ReactorImpl ) StartList() error {
    if err := ri.checkActive( "StartList" ); err != nil { return err }
    if err := ri.checkValue( "list start" ); err != nil { return err }
    ri.push( "list" )
    return nil
}

func ( ri *ReactorImpl ) StartField( fld *mg.Identifier ) error {
    if err := ri.checkActive( "StartField" ); err != nil { return err }
    if ok := ri.stack.Len() > 0; ok {
        elt := ri.stack.Front().Value
        switch v := elt.( type ) {
        case string: 
            tmpl := "Expected list value but got start of field '%s'"
            return Errorf( tmpl, fld )
        case *reactorImplMap: return v.startField( fld )
        default: panic( errorf( "Invalid stack element: %T", elt ) )
        }
    }
    errLoc := fmt.Sprintf( "field '%s'", fld )
    return getTopLevelStructStartExpectError( errLoc )
}

func ( ri *ReactorImpl ) Value() error {
    if err := ri.checkActive( "Value" ); err != nil { return err }
    if err := ri.checkValue( "value" ); err != nil { return err }
    return nil
}

func ( ri *ReactorImpl ) End() error {
    if err := ri.checkActive( "End" ); err != nil { return err }
    if ri.stack.Len() == 0 { return getTopLevelStructStartExpectError( "end" ) }
    elt := ri.stack.Remove( ri.stack.Front() )
    switch v := elt.( type ) {
    case *reactorImplMap: if err := v.end(); err != nil { return err }
    case string: {} // list -- end() is always okay
    default: panic( errorf( "Unexpected stack element: %T", elt ) )
    }
    ri.done = ri.stack.Len() == 0
    return nil
}

type valAccumulator interface {
    valueReady( val mg.Value ) 
    end() ( mg.Value, error )
}

type mapAcc struct {
    arr []interface{} // alternating key, val to be passed to mg.MustSymbolMap
}

func newMapAcc() *mapAcc {
    return &mapAcc{ arr: make( []interface{}, 0, 8 ) }
}

func ( ma *mapAcc ) end() ( mg.Value, error ) { 
    res, err := mg.CreateSymbolMap( ma.arr... )
    if err == nil { return res, nil } 
    return nil, Errorf( "Invalid fields: %s", err.Error() )
}

func ( ma *mapAcc ) startField( fld *mg.Identifier ) {
    ma.arr = append( ma.arr, fld )
}

func ( ma *mapAcc ) valueReady( mv mg.Value ) { ma.arr = append( ma.arr, mv ) }

type structAcc struct {
    typ mg.TypeReference
    flds *mapAcc
}

func newStructAcc( typ mg.TypeReference ) *structAcc {
    return &structAcc{ typ: typ, flds: newMapAcc() }
}

func ( sa *structAcc ) end() ( mg.Value, error ) {
    flds, err := sa.flds.end()
    if err != nil { return nil, err }
    return &mg.Struct{ Type: sa.typ, Fields: flds.( *mg.SymbolMap ) }, nil
}

func ( sa *structAcc ) valueReady( mv mg.Value ) { sa.flds.valueReady( mv ) }

type listAcc struct {
    vals []mg.Value
}

func newListAcc() *listAcc {
    return &listAcc{ make( []mg.Value, 0, 4 ) }
}

func ( la *listAcc ) end() ( mg.Value, error ) { 
    return mg.NewList( la.vals ), nil
}

func ( la *listAcc ) valueReady( mv mg.Value ) {
    la.vals = append( la.vals, mv )
}

type StructBuilder struct {
    s *mg.Struct
    accs *list.List
    impl *ReactorImpl
}

func NewStructBuilder() *StructBuilder {
    return &StructBuilder{ accs: &list.List{}, impl: NewReactorImpl() }
}

func ( sb *StructBuilder ) Ready() bool { return sb.s != nil }

func ( sb *StructBuilder ) pushAcc( acc valAccumulator ) {
    sb.accs.PushFront( acc )
}

func ( sb *StructBuilder ) peekAcc() ( valAccumulator, bool ) {
    if sb.accs.Len() == 0 { return nil, false }
    return sb.accs.Front().Value.( valAccumulator ), true
}

func ( sb *StructBuilder ) popAcc() valAccumulator {
    res, ok := sb.peekAcc()
    if ! ok { panic( errorf( "popAcc() called on empty stack" ) ) }
    sb.accs.Remove( sb.accs.Front() )
    return res
}

func ( sb *StructBuilder ) valueReady( val mg.Value ) {
    if acc, ok := sb.peekAcc(); ok {
        acc.valueReady( val )
    } else { sb.s = val.( *mg.Struct ) }
}

// Panics if result of Ready() is false
func ( sb *StructBuilder ) GetStruct() *mg.Struct {
    if sb.Ready() { return sb.s }
    panic( Errorf( "Attempt to access incomplete struct" ) )
}

func ( sb *StructBuilder ) StartStruct( typ mg.TypeReference ) error {
    if err := sb.impl.StartStruct(); err != nil { return err }
    sb.pushAcc( newStructAcc( typ ) )
    return nil
}

func ( sb *StructBuilder ) StartMap() error {
    if err := sb.impl.StartMap(); err != nil { return err }
    sb.pushAcc( newMapAcc() )
    return nil
}

func ( sb *StructBuilder ) StartList() error {
    if err := sb.impl.StartList(); err != nil { return err }
    sb.pushAcc( newListAcc() )
    return nil
}

func ( sb *StructBuilder ) StartField( fld *mg.Identifier ) error {
    if err := sb.impl.StartField( fld ); err != nil { return err }
    acc, ok := sb.peekAcc()
    if ok {
        var ma *mapAcc
        switch v := acc.( type ) {
        case *mapAcc: ma, ok = v, true
        case *structAcc: ma, ok = v.flds, true
        default: ok = false
        }
        if ok { ma.startField( fld ) }
    }
    return nil
}

func ( sb *StructBuilder ) End() error {
    if err := sb.impl.End(); err != nil { return err }
    acc := sb.popAcc()
    if val, err := acc.end(); err == nil {
        sb.valueReady( val )
    } else { return err }
    return nil
}

func ( sb *StructBuilder ) Value( mv mg.Value ) error {
    if err := sb.impl.Value(); err != nil { return err }
    sb.valueReady( mv ) 
    return nil
}

func visitSymbolMap( 
    m *mg.SymbolMap, callStart bool, rct Reactor ) error {
    if callStart {
        if err := rct.StartMap(); err != nil { return err }
    }
    err := m.EachPairError( func( fld *mg.Identifier, val mg.Value ) error {
        if err := rct.StartField( fld ); err != nil { return err }
        return visitValue( val, rct )
    })
    if err != nil { return err }
    return rct.End()
}

func visitStruct( ms *mg.Struct, rct Reactor ) ( err error ) {
    if err = rct.StartStruct( ms.Type ); err != nil { return }
    return visitSymbolMap( ms.Fields, false, rct )
}

func visitList( ml *mg.List, rct Reactor ) ( err error ) {
    if err = rct.StartList(); err != nil { return }
    for _, val := range ml.Values() {
        if err = visitValue( val, rct ); err != nil { return }
    }
    return rct.End()
}

func visitValue( mv mg.Value, rct Reactor ) error {
    switch v := mv.( type ) {
    case *mg.Struct: return visitStruct( v, rct )
    case *mg.SymbolMap: return visitSymbolMap( v, true, rct )
    case *mg.List: return visitList( v, rct )
    }
    return rct.Value( mv )
}

func VisitStruct( ms *mg.Struct, rct Reactor ) error {
    return visitValue( ms, rct )
}

type Codec interface {
    EncoderTo( w io.Writer ) Reactor
    DecodeFrom( rd io.Reader, rct Reactor ) error
}

func Encode( ms *mg.Struct, cdc Codec, w io.Writer ) error {
    rct := cdc.EncoderTo( w )
    return VisitStruct( ms, rct )
}

func EncodeBytes( ms *mg.Struct, cdc Codec ) ( []byte, error ) {
    buf := &bytes.Buffer{}
    if err := Encode( ms, cdc, buf ); err != nil { return nil, err }
    return buf.Bytes(), nil
}

func Decode( cdc Codec, rd io.Reader ) ( *mg.Struct, error ) {
    rct := NewStructBuilder()
    if err := cdc.DecodeFrom( rd, rct ); err != nil { return nil, err }
    return rct.GetStruct(), nil
}

func DecodeBytes( cdc Codec, buf []byte ) ( *mg.Struct, error ) {
    return Decode( cdc, bytes.NewBuffer( buf ) )
}

type CodecError struct {
    msg string
}

func ( e *CodecError ) Error() string { return e.msg }

func Error( msg string ) *CodecError { return &CodecError{ msg } }

func Errorf( msg string, args ...interface{} ) *CodecError {
    return Error( fmt.Sprintf( msg, args... ) )
}

func regKeyFor( id *mg.Identifier ) string { return id.ExternalForm() }

type CodecRegistrationError struct { msg string }
func ( e *CodecRegistrationError ) Error() string { return e.msg }

type CodecRegistration struct {
    Codec Codec
    Id *mg.Identifier
    Source string
}

func ( r *CodecRegistration ) key() string { return regKeyFor( r.Id ) }

var registry = make( map[ string ]*CodecRegistration )

func GetCodecById( id *mg.Identifier ) Codec {
    if reg := registry[ regKeyFor( id ) ]; reg != nil { return reg.Codec }
    return nil
}

func MustCodecById( id *mg.Identifier ) Codec {
    if cdc := GetCodecById( id ); cdc != nil { return cdc }
    panic( fmt.Errorf( "No such codec: %s", id ) )
}

func RegisterCodec( reg *CodecRegistration ) error {
    key := reg.key()
    if prev, ok := registry[ key ]; ok {
        msg := "Codec %q already registered by %q"
        return &CodecRegistrationError{ fmt.Sprintf( msg, key, prev.Source ) }
    }
    registry[ key ] = reg
    return nil
}
