package pipeline

import (
    "fmt"
)

type Visitor func( elt interface{} )

type Initializer interface {
    InitializePipeline( pip *Pipeline )
}

type Pipeline struct {
    elts []interface{}
}

func NewPipeline() *Pipeline {
    return &Pipeline{ elts: make( []interface{}, 0, 4 ) }
}

func ( pip *Pipeline ) Len() int { return len( pip.elts ) }

func ( pip *Pipeline ) Get( i int ) interface{} { 
    if l := pip.Len(); i < 0 || i >= l { 
        tmpl := "index %d is out of range (pipeline length is %d)"
        panic( fmt.Errorf( tmpl, i, l ) )
    }
    return pip.elts[ i ]
}

func ( pip *Pipeline ) Add( elt interface{} ) {
    if pi, ok := elt.( Initializer ); ok { pi.InitializePipeline( pip ) }
    pip.elts = append( pip.elts, elt )
}

func ( pip *Pipeline ) VisitReverse( f Visitor ) {
    for i := pip.Len(); i > 0; i-- { f( pip.Get( i - 1 ) ) }
}
