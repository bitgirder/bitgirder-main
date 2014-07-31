package bind

import (
    "bitgirder/objpath"
    "fmt"
    mg "mingle"
    mgRct "mingle/reactor"
    "time"
//    "log"
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

func bindErrorFactory( path objpath.PathNode, msg string ) error {
    return NewBindError( path, msg )
}

// could make this public if needed
func addPrimBindings( reg *BindRegistry ) {
    addPrim := func( qn *mg.QualifiedTypeName, f mgRct.BuildValueOkFunction ) {
        bf := mgRct.NewFunctionsBuilderFactory()
        bf.ErrorFunc = bindErrorFactory
        bf.ValueFunc = f
        reg.MustAddValue( qn, bf )
    }
    addPrim(
        mg.QnameNull,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if _, ok := ve.Val.( *mg.Null ); ok { return nil, nil, true }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameBoolean, 
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Boolean ); ok {
                return bool( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameBuffer,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Buffer ); ok {
                return []byte( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameString,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.String ); ok {
                return string( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameInt32,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Int32 ); ok {
                return int32( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameUint32,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Uint32 ); ok {
                return uint32( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameFloat32,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Float32 ); ok {
                return float32( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameInt64,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Int64 ); ok {
                return int64( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameUint64,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Uint64 ); ok {
                return uint64( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameFloat64,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Float64 ); ok {
                return float64( v ), nil, true
            }
            return nil, nil, false
        },
    )
    addPrim(
        mg.QnameTimestamp,
        func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
            if v, ok := ve.Val.( mg.Timestamp ); ok {
                return time.Time( v ), nil, true
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
    res.ErrorFunc = bindErrorFactory
    res.ValueFunc = func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
        qn := mg.TypeOf( ve.Val ).( *mg.AtomicTypeReference ).Name
        if bf, ok := reg.m.GetOk( qn ); ok {
            res, err := bf.( mgRct.BuilderFactory ).BuildValue( ve )
            return res, err, true
        }
        return nil, nil, false
    }
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        
        if bf, ok := reg.m.GetOk( sse.Type ); ok {
            res, err := bf.( mgRct.BuilderFactory ).StartStruct( sse )
            return res, err
        }
        return nil, nil
    }
    return res
}
