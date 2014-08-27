package io

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "io"
)

type BinWriter struct { *mg.BinWriter }

func NewWriter( w io.Writer ) *BinWriter { 
    return &BinWriter{ mg.NewWriter( w ) }
}

type writeReactor struct { *BinWriter }

func ( w writeReactor ) startStruct( qn *mg.QualifiedTypeName ) error {
    if err := w.WriteTypeCode( mg.IoTypeCodeStruct ); err != nil { return err }
    return w.WriteQualifiedTypeName( qn )
}

func ( w writeReactor ) startField( fld *mg.Identifier ) error {
    if err := w.WriteTypeCode( mg.IoTypeCodeField ); err != nil { return err }
    return w.WriteIdentifier( fld )
}

func ( w writeReactor ) startList( lse *mgRct.ListStartEvent ) error { 
    if err := w.WriteTypeCode( mg.IoTypeCodeList ); err != nil { return err }
    if err := w.WriteListTypeReference( lse.Type ); err != nil { return err }
    return nil
}

func ( w writeReactor ) startMap() error { 
    return w.WriteTypeCode( mg.IoTypeCodeSymMap )
}

func ( w writeReactor ) value( val mg.Value ) error {
    return w.WriteScalarValue( val )
}

func ( w writeReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return w.value( v.Val )
    case *mgRct.MapStartEvent: return w.startMap()
    case *mgRct.StructStartEvent: return w.startStruct( v.Type )
    case *mgRct.ListStartEvent: return w.startList( v )
    case *mgRct.FieldStartEvent: return w.startField( v.Field )
    case *mgRct.EndEvent: return w.WriteTypeCode( mg.IoTypeCodeEnd )
    }
    panic( libErrorf( "unhandled event type: %T", ev ) )
}

func ( w *BinWriter ) AsReactor() mgRct.ReactorEventProcessor { 
    return writeReactor{ w } 
}

func ( w *BinWriter ) WriteValue( val mg.Value ) ( err error ) {
    return mgRct.VisitValue( val, w.AsReactor() )
}

type BinReader struct {
    *mg.BinReader
}

func NewReader( r io.Reader ) *BinReader {
    return &BinReader{ mg.NewReader( r ) }
}

func ( r *BinReader ) readScalarValue( 
    tc mg.IoTypeCode, rep mgRct.ReactorEventProcessor ) error {

    val, err := r.ReadScalarValue( tc )
    if err != nil { return err }
    return rep.ProcessEvent( mgRct.NewValueEvent( val ) )
}

func ( r *BinReader ) readMapFields( rep mgRct.ReactorEventProcessor ) error {
    for {
        tc, err := r.ReadTypeCode()
        if err != nil { return err }
        switch tc {
        case mg.IoTypeCodeEnd: return rep.ProcessEvent( mgRct.NewEndEvent() )
        case mg.IoTypeCodeField:
            id, err := r.ReadIdentifier()
            if err == nil { 
                err = rep.ProcessEvent( mgRct.NewFieldStartEvent( id ) ) 
            }
            if err != nil { return err }
            if err := r.implReadValue( rep ); err != nil { return err }
        default: return r.IoErrorf( "Unexpected map pair code: 0x%02x", tc )
        }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( r *BinReader ) readSymbolMap( rep mgRct.ReactorEventProcessor ) error {
    if err := rep.ProcessEvent( mgRct.NewMapStartEvent() ); err != nil {
        return err 
    }
    return r.readMapFields( rep )
}

func ( r *BinReader ) readStruct( rep mgRct.ReactorEventProcessor ) error {
    if qn, err := r.ReadQualifiedTypeName(); err == nil {
        ev := mgRct.NewStructStartEvent( qn )
        if err = rep.ProcessEvent( ev ); err != nil { return err }
    } else { return err }
    return r.readMapFields( rep )
}

func ( r *BinReader ) readListHeader( rep mgRct.ReactorEventProcessor ) error {
    if typ, err := r.ReadListTypeReference(); err == nil {
        lse := mgRct.NewListStartEvent( typ )
        if err = rep.ProcessEvent( lse ); err != nil { return err }
    } else { return err }
    return nil
}

func ( r *BinReader ) readListValues( rep mgRct.ReactorEventProcessor ) error {
    for {
        tc, err := r.PeekTypeCode()
        if err != nil { return err }
        if tc == mg.IoTypeCodeEnd {
            if _, err = r.ReadTypeCode(); err != nil { return err }
            return rep.ProcessEvent( mgRct.NewEndEvent() )
        } else { 
            if err = r.implReadValue( rep ); err != nil { return err } 
        }
    }
    panic( libErrorf( "Unreachable" ) )
}

func ( r *BinReader ) readList( rep mgRct.ReactorEventProcessor ) error {
    if err := r.readListHeader( rep ); err != nil { return err }
    return r.readListValues( rep )
}

func ( r *BinReader ) implReadValue( rep mgRct.ReactorEventProcessor ) error {
    tc, err := r.ReadTypeCode()
    if err != nil { return err }
    switch tc {
    case mg.IoTypeCodeNull, mg.IoTypeCodeString, mg.IoTypeCodeBuffer, 
         mg.IoTypeCodeTimestamp, mg.IoTypeCodeInt32, mg.IoTypeCodeInt64, 
         mg.IoTypeCodeUint32, mg.IoTypeCodeUint64, mg.IoTypeCodeFloat32,
         mg.IoTypeCodeFloat64, mg.IoTypeCodeBool, mg.IoTypeCodeEnum:
        return r.readScalarValue( tc, rep )
    case mg.IoTypeCodeSymMap: return r.readSymbolMap( rep )
    case mg.IoTypeCodeStruct: return r.readStruct( rep )
    case mg.IoTypeCodeList: return r.readList( rep )
    default: return r.IoErrorf( "unrecognized value code: 0x%02x", tc )
    }
    panic( libErrorf( "unreachable" ) )
}

func ( r *BinReader ) ReadReactorValue( 
    rep mgRct.ReactorEventProcessor ) error {

    return r.implReadValue( rep )
}

func ( r *BinReader ) ReadValue() ( mg.Value, error ) {
    vb := mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    pip := mgRct.InitReactorPipeline( vb )
    err := r.ReadReactorValue( pip )
    if err != nil { return nil, err }
    return vb.GetValue().( mg.Value ), nil
}
