package reactor

import (
    mg "mingle"
    "bitgirder/pipeline"
//    "log"
)

type PointerCheckReactor struct {
    m map[ mg.PointerId ] bool
}

func NewPointerCheckReactor() *PointerCheckReactor {
    return &PointerCheckReactor{ m: make( map[ mg.PointerId ] bool ) }
}

func ( r *PointerCheckReactor ) checkReference( 
    re *ValueReferenceEvent ) error {

    if _, ok := r.m[ re.Id ]; ok { return nil }
    if re.Id == mg.PointerIdNull {
        return rctError( re.GetPath(), "attempt to reference null pointer" )
    }
    return rctErrorf( re.GetPath(), "unrecognized reference: %s", re.Id )
}

func ( r *PointerCheckReactor ) checkAlloc( 
    ev ReactorEvent, id mg.PointerId ) error {

    if id == mg.PointerIdNull { return nil }
    if _, ok := r.m[ id ]; ok {
        tmpl := "attempt to redefine reference: %s"
        return rctErrorf( ev.GetPath(), tmpl, id )
    }
    r.m[ id ] = true
    return nil
}

func ( r *PointerCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueAllocationEvent: return r.checkAlloc( v, v.Id )
    case *ListStartEvent: return r.checkAlloc( v, v.Id )
    case *MapStartEvent: return r.checkAlloc( v, v.Id )
    case *ValueReferenceEvent: return r.checkReference( v )
    }
    return nil
}

func EnsurePointerCheckReactor( pip *pipeline.Pipeline ) {
    var chk *PointerCheckReactor
    pip.VisitReverse( func( elt interface{} ) {
        if chk == nil { chk, _ = elt.( *PointerCheckReactor ) }
    })
    if chk == nil { pip.Add( NewPointerCheckReactor() ) }
}
