package bind

import (
    "bitgirder/objpath"
    "fmt"
    mg "mingle"
    mgRct "mingle/reactor"
)

type BindError struct {
    Path objpath.PathNode
    Message string
}

func ( e *BindError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewBindError( path objpath.PathNode, msg string ) *BindError {
    return &BindError{ Path: path, Message: msg }
}

func NewBindErrorf( 
    path objpath.PathNode, tmpl string, argv ...interface{} ) *BindError {

    return NewBindError( path, fmt.Sprintf( tmpl, argv... ) )
}

var DomainDefault = mg.NewIdentifierUnsafe( []string{ "default" } )

type BindRegistry struct {
    m *mg.QnameMap 
}

func NewBindRegistry() *BindRegistry { 
    return &BindRegistry{ m: mg.NewQnameMap() }
}

func ( reg *BindRegistry ) MustAddValue( 
    qn *mg.QualifiedTypeName, bf mgRct.BuilderFactory ) {

    if reg.m.HasKey( qn ) {
        panic( libErrorf( "registry already binds type: %s", qn ) )
    }
    reg.m.Put( qn, bf )
}

// could make this public if needed
func addPrimBindings( reg *BindRegistry ) {
    addPrim := func( qn *mg.QualifiedTypeName, f mgRct.BuildValueOkFunction ) {
        bf := mgRct.NewFunctionsBuilderFactory()
        bf.ValueFunc = f
        reg.MustAddValue( qn, bf )
    }
    addPrim(
        mg.QnameBoolean, 
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if b, ok := ve.Val.( mg.Boolean ); ok {
                return bool( b ), nil, true
            }
            return nil, nil, false
        },
    )
}

var regsByDomain *mg.IdentifierMap = mg.NewIdentifierMap()

func init() {
    reg := NewBindRegistry()
    regsByDomain.Put( DomainDefault, reg )
    addPrimBindings( reg )
}

func BindRegistryForDomain( domain *mg.Identifier ) *BindRegistry {
    if reg, ok := regsByDomain.GetOk( domain ); ok { 
        return reg.( *BindRegistry )
    }
    return nil
}

func NewBindBuilderFactory( reg *BindRegistry ) mgRct.BuilderFactory {
    res := mgRct.NewFunctionsBuilderFactory()
    res.ErrorFunc = func( path objpath.PathNode, msg string ) error {
        return NewBindError( path, msg )
    }
    res.ValueFunc = func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
        qn := mg.TypeOf( ve.Val ).( *mg.AtomicTypeReference ).Name
        if bf, ok := reg.m.GetOk( qn ); ok {
            res, err := bf.( mgRct.BuilderFactory ).BuildValue( ve )
            return res, err, true
        }
        return nil, nil, false
    }
    return res
}
