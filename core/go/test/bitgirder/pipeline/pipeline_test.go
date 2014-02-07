package pipeline

import (
    "testing"
    "container/list"
    "reflect"
)

type elt1 struct { i int }

type elt2 struct { i int }

func ( e elt2 ) InitializePipeline( pip *Pipeline ) { pip.Add( elt1{ e.i } ) }

type elt3 struct { i int }

func ( e elt3 ) InitializePipeline( pip *Pipeline ) { pip.Add( elt2{ e.i } ) }

type elt4 struct { i int }

func ( e elt4 ) InitializePipeline( pip *Pipeline ) { pip.Add( elt3{ e.i } ) }

func assertPipeline( t *testing.T, pip *Pipeline, elts ...interface{} ) {
    if l := pip.Len(); l != len( elts ) {
        t.Fatalf( "expected len %d but got %d", len( elts ), l )
    }
    for i, e := 0, pip.Len(); i < e; i++ {
        expct := elts[ i ]
        act := pip.Get( i )
        if ! reflect.DeepEqual( expct, act ) {
            t.Fatalf( "at %d, expct != act: %v != %v", i, expct, act )
        }
    }
}

func TestPipelineEmpty( t *testing.T ) {
    assertPipeline( t, NewPipeline() )
}

func TestPipelineBasic( t *testing.T ) {
    pip := NewPipeline()
    assertPipeline( t, pip )
    pip.Add( elt1{ 1 } )
    assertPipeline( t, pip, elt1{ 1 } )
    pip.Add( elt2{ 2 } )
    assertPipeline( t, pip, elt1{ 1 }, elt1{ 2 }, elt2{ 2 } )
    pip.Add( elt4{ 3 } )
    assertPipeline( t, pip, 
        elt1{ 1 }, 
        elt1{ 2 }, elt2{ 2 }, 
        elt1{ 3 }, elt2{ 3 }, elt3{ 3 }, elt4{ 3 },
    )
    l := &list.List{}
    pip.VisitReverse( func( e interface{} ) { l.PushFront( e ) } )
    vis := []interface{}{}
    for l.Len() > 0 { vis = append( vis, l.Remove( l.Front() ) ) }
    assertPipeline( t, pip, vis... )
}
