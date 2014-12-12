package bind

import (
    "bitgirder/objpath"
    "fmt"
    "mingle/parser"
    mg "mingle"
    mgRct "mingle/reactor"
    "time"
    "reflect"
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

func bindErrorFactory( path objpath.PathNode, msg string ) error {
    return NewBindError( path, msg )
}

type VisitContext struct {
    Destination mgRct.EventProcessor
    Path objpath.PathNode
    BindContext *BindContext
}

func ( vc VisitContext ) EventSender() mgRct.EventSender {
    return mgRct.EventSenderForReactor( vc.Destination )
}

type VisitBodyFunc func() error

func VisitStruct( 
    vc VisitContext, typ *mg.QualifiedTypeName, f VisitBodyFunc ) error {

    es := vc.EventSender()
    if err := es.StartStruct( typ ); err != nil { return err }
    if err := f(); err != nil { return err }
    if err := es.End(); err != nil { return err }
    return nil
}

func VisitFieldFunc(
    vc VisitContext, 
    fld *mg.Identifier, 
    val interface{},
    f VisitValueFunc ) error {

    if err := vc.EventSender().StartField( fld ); err != nil { return err }
    return f( val, vc )
}

func VisitFieldValue( 
    vc VisitContext, fld *mg.Identifier, val interface{} ) error {

    return VisitFieldFunc( vc, fld, val, VisitValue )
}

func VisitList( vc VisitContext,
                lt *mg.ListTypeReference,
                body VisitBodyFunc ) error {

    es := vc.EventSender()
    if err := es.StartList( lt ); err != nil { return err }
    if err := body(); err != nil { return err }
    return es.End()
}

func VisitListFunc(
    vc VisitContext, 
    lt *mg.ListTypeReference, 
    listLen int, 
    f func( i int ) error ) error {

    return VisitList( vc, lt, func() error {
        for i := 0; i < listLen; i++ {
            if err := f( i ); err != nil { return err }
        }
        return nil
    })
}

func VisitListValue(
    vc VisitContext,
    lt *mg.ListTypeReference,
    listLen int,
    f func( i int ) interface{} ) error {

    return VisitListFunc( vc, lt, listLen, func( i int ) error {
        return VisitValue( f( i ), vc )
    })
}

type VisitValueFunc func( val interface{}, vc VisitContext ) error
type VisitValueOkFunc func( val interface{}, vc VisitContext ) ( error, bool )

type Registry struct {
    m *mg.QnameMap 
    visitors []VisitValueOkFunc
}

func NewRegistry() *Registry { 
    return &Registry{ 
        m: mg.NewQnameMap(),
        visitors: make( []VisitValueOkFunc, 0, 4 ),
    }
}

func ( reg *Registry ) BuilderFactoryForName( 
    nm *mg.QualifiedTypeName ) ( mgRct.BuilderFactory, bool ) {
        
    if v, ok := reg.m.GetOk( nm ); ok { 
        return v.( mgRct.BuilderFactory ), true
    }
    if nm.Equals( mg.QnameValue ) { return NewBuilderFactory( reg ), true }
    return nil, false
}

func ( reg *Registry ) BuilderFactoryForType( 
    typ mg.TypeReference ) ( mgRct.BuilderFactory, bool ) {

    if at, ok := typ.( *mg.AtomicTypeReference ); ok {
        return reg.BuilderFactoryForName( at.Name() )
    }
    return nil, false
}

func ( reg *Registry ) MustBuilderFactoryForType( 
    typ mg.TypeReference ) mgRct.BuilderFactory {

    if res, ok := reg.BuilderFactoryForType( typ ); ok { return res }
    panic( libErrorf( "no builder factory for type: %s", typ ) )
}

func ( reg *Registry ) MustAddValue( 
    qn *mg.QualifiedTypeName, bf mgRct.BuilderFactory ) {

    if reg.m.HasKey( qn ) {
        panic( libErrorf( "registry already binds type: %s", qn ) )
    }
    reg.m.Put( qn, bf )
}

func ( reg *Registry ) AddVisitValueOkFunc( f VisitValueOkFunc ) {
    reg.visitors = append( reg.visitors, f )
}

func NewFunctionsBuilderFactory() *mgRct.FunctionsBuilderFactory {
    res := mgRct.NewFunctionsBuilderFactory()
    res.ErrorFactory = bindErrorFactory
    return res
}

func visitPrimValueOk( val interface{}, vc VisitContext ) ( error, bool ) {
    switch v := val.( type ) {
    case bool, []byte, string, int32, int64, uint32, uint64, float32, float64,
         time.Time, nil: 
        return visitPrimValueOk( mg.MustValue( v ), vc )
    case mg.Value: 
        return mgRct.VisitValuePath( v, vc.Destination, vc.Path ), true
    }
    return nil, false
}

// could make this public if needed
func addPrimBindings( reg *Registry ) {
    addPrim := func( qn *mg.QualifiedTypeName, f mgRct.BuildValueOkFunction ) {
        bf := NewFunctionsBuilderFactory()
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
    reg.AddVisitValueOkFunc( visitPrimValueOk )
}

var regsByDomain *mg.IdentifierMap = mg.NewIdentifierMap()

func init() {
    reg := NewRegistry()
    regsByDomain.Put( DomainDefault, reg )
    addPrimBindings( reg )
}

func RegistryForDomain( domain *mg.Identifier ) *Registry {
    if reg, ok := regsByDomain.GetOk( domain ); ok { 
        return reg.( *Registry )
    }
    return nil
}

func MustRegistryForDomain( domain *mg.Identifier ) *Registry {
    if res := RegistryForDomain( domain ); res != nil { return res }
    panic( libErrorf( "no registry for domain: %s", domain ) )
}

func newBuilderFactory( reg *Registry ) *mgRct.FunctionsBuilderFactory {
    res := NewFunctionsBuilderFactory()
    res.ValueFunc = func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
        qn := mg.TypeOf( ve.Val ).( *mg.AtomicTypeReference ).Name()
        if bf, ok := reg.BuilderFactoryForName( qn ); ok {
            res, err := bf.( mgRct.BuilderFactory ).BuildValue( ve )
            return res, err, true
        }
        return nil, nil, false
    }
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        
        if bf, ok := reg.BuilderFactoryForName( sse.Type ); ok {
            res, err := bf.( mgRct.BuilderFactory ).StartStruct( sse )
            return res, err
        }
        return nil, nil
    }
    return res
}

// public frontend to newBuilderFactory that allows us to have the return type
// be simply mgRct.BuilderFactory while retaining type safety for internal
// callers of newBuilderFactory
func NewBuilderFactory( reg *Registry ) mgRct.BuilderFactory {
    return newBuilderFactory( reg )
}

type opaqueMapBuilder struct {
    m map[ string ] interface{}
    f mgRct.BuilderFactory // the parent opaque builder factory
}

func ( b opaqueMapBuilder ) ProduceValue( 
    _ *mgRct.EndEvent ) ( interface{}, error ) {

    return b.m, nil
}

func ( b opaqueMapBuilder ) StartField( 
    fse *mgRct.FieldStartEvent ) ( mgRct.BuilderFactory, error ) {

    return b.f, nil
}

func ( b opaqueMapBuilder ) SetValue( 
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    b.m[ fld.ExternalForm() ] = val
    return nil
}

type opaqueListBuilder struct {
    l []interface{}
    f mgRct.BuilderFactory // parent opaque builder factory
    reg *Registry
}

func ( b *opaqueListBuilder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    b.l = append( b.l, val )
    return nil
}

func ( b *opaqueListBuilder ) NextBuilderFactory() mgRct.BuilderFactory {
    return b.f
}

func ( b *opaqueListBuilder ) ProduceValue(
    _ *mgRct.EndEvent ) ( interface{}, error ) {

    return b.l, nil
}

func NewOpaqueValueFactory( reg *Registry ) mgRct.BuilderFactory {
    res := newBuilderFactory( reg )
    res.MapFunc = func( 
        _ *mgRct.MapStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        
        m := make( map[ string ] interface{}, 4 )
        return opaqueMapBuilder{ f: res, m: m }, nil
    }
    res.ListFunc = func( 
        _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {

        l := make( []interface{}, 0, 4 )
        return &opaqueListBuilder{ f: res, l: l }, nil
    }
    return res
}

func NewBuildReactor( bf mgRct.BuilderFactory ) *mgRct.BuildReactor {
    res := mgRct.NewBuildReactor( bf )
    res.ErrorFactory = bindErrorFactory
    return res
}

var (
    NewFunctionsListBuilder = mgRct.NewFunctionsListBuilder
    NewFunctionsFieldSetBuilder = mgRct.NewFunctionsFieldSetBuilder
)

type SerialFormat int

const (
    SerialFormatDefault = SerialFormat( iota )
    SerialFormatBinary 
    SerialFormatText
)

type SerialOptions struct {
    Identifiers mg.IdentifierFormat
    Format SerialFormat
}

func NewSerialOptions() *SerialOptions {
    return &SerialOptions{
        Identifiers: mg.LcHyphenated,
        Format: SerialFormatDefault,
    }
}

type BindContext struct {
    Registry *Registry
    SerialOptions *SerialOptions
}

func NewBindContext( reg *Registry ) *BindContext {
    return &BindContext{ 
        Registry: reg,
        SerialOptions: NewSerialOptions(),
    }
}

type VisitError struct {
    Location objpath.PathNode
    Message string
}

func NewVisitError( path objpath.PathNode, msg string ) *VisitError {
    return &VisitError{ Location: path, Message: msg }
}

func NewVisitErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *VisitError {

    return NewVisitError( path, fmt.Sprintf( tmpl, args... ) )
}

func ( e *VisitError ) Error() string {
    return mg.FormatError( e.Location, e.Message )
}

type ValueVisitor interface {

    VisitValue( vc VisitContext ) error
}

// could make this public if needed
func visitValueOk( val interface{}, vc VisitContext ) ( bool, error ) {
    if vv, ok := val.( ValueVisitor ); ok { return true, vv.VisitValue( vc ) }
    for _, f := range vc.BindContext.Registry.visitors {
        if err, ok := f( val, vc ); ok { return true, err }
    }
    return false, nil
}

func errForUnknownVisitType( val interface{}, vc VisitContext ) error {
    return NewVisitErrorf( vc.Path, "unknown type for visit: %T", val )
}

func VisitValue( val interface{}, vc VisitContext ) error {
    if ok, err := visitValueOk( val, vc ); ok { return err }
    return errForUnknownVisitType( val, vc )
}

func visitPtrValueOpaque( val interface{}, vc VisitContext ) error {
    ptrVal := reflect.ValueOf( val )
    return VisitValueOpaque( ptrVal.Elem().Interface(), vc )
}

func visitSliceValueOpaque( val interface{}, vc VisitContext ) error {
    slc := reflect.ValueOf( val )
    es := vc.EventSender()
    if err := es.StartList( mg.TypeOpaqueList ); err != nil { return err }
    lp := objpath.StartList( vc.Path )
    for i, e := 0, slc.Len(); i < e; i++ {
        vc2 := vc // shallow copy
        vc2.Path = lp
        elt := slc.Index( i ).Interface()
        if err := VisitValueOpaque( elt, vc2 ); err != nil { return err }
        lp = lp.Next()
    }
    return es.End()
}

//func visitStructValueOpaque( val interface{}, vc VisitContext ) error {
//    s := reflect.ValueOf( val )
//    if ! s.CanAddr() { return errForUnknownVisitType( val, vc ) }
//    return VisitValue( s.Addr().Interface(), vc )
//}

func visitMapOpaque( m map[ string ] interface{}, vc VisitContext ) error {
    es := vc.EventSender()
    if err := es.StartMap(); err != nil { return err }
    for k, v := range m {
        id, err := parser.ParseIdentifier( k )
        if err != nil { return NewBindError( vc.Path, err.Error() ) }
        if err = es.StartField( id ); err != nil { return err }
        vc2 := vc // shallow copy vc
        vc2.Path = objpath.Descend( vc2.Path, id )
        if err = VisitValueOpaque( v, vc2 ); err != nil { return err }
    }
    return es.End()
}

func VisitValueOpaque( val interface{}, vc VisitContext ) error {
    if ok, err := visitValueOk( val, vc ); ok { return err }
    if val == nil { return VisitValue( mg.NullVal, vc ) }
    if m, ok := val.( map[ string ] interface{} ); ok {
        return visitMapOpaque( m, vc )
    }
    switch reflect.TypeOf( val ).Kind() {
    case reflect.Ptr: return visitPtrValueOpaque( val, vc )
    case reflect.Slice: return visitSliceValueOpaque( val, vc )
//    case reflect.Struct: return visitStructValueOpaque( val, vc )
    }
    return errForUnknownVisitType( val, vc )
}
