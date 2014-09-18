package bind

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

func TypedFieldStarter( 
    reg *Registry, typ mg.TypeReference ) mgRct.StartFieldFunction {

    return func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
        if res, ok := reg.BuilderFactoryForType( typ ); ok {
            return res.( mgRct.BuilderFactory ), nil
        }
        return nil, nil
    }
}

type CheckedFieldSetValFunction func( obj, fldVal interface{} )

type CheckedFieldStartFunction func( reg *Registry ) mgRct.BuilderFactory

type CheckedFieldSetter struct {

    Field *mg.Identifier

    Type mg.TypeReference
    StartField CheckedFieldStartFunction

    Assign CheckedFieldSetValFunction
}

func ( fld *CheckedFieldSetter ) fieldStarter( 
    reg *Registry ) mgRct.StartFieldFunction {

    if sf := fld.StartField; sf != nil { 
        return func( _ objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return sf( reg ), nil
        }
    }
    return TypedFieldStarter( reg, fld.Type )
}

func ( fld *CheckedFieldSetter ) mustValidate() {
    set := 0
    if fld.StartField != nil { set++ }
    if fld.Type != nil { set++ }
    nm := fld.Field
    switch set {
    case 0: panic( libErrorf( "no way to start field for %s", nm ) )
    case 1: return
    default: panic( libErrorf( "multiple field starters for %s", nm ) )
    }
}

func validateSetters( setters []*CheckedFieldSetter ) {
    for _, s := range setters { s.mustValidate() }
}

func AddCheckedField(
    fsb *mgRct.FunctionsFieldSetBuilder,
    reg *Registry,
    fld *CheckedFieldSetter ) {

    fsb.RegisterField(
        fld.Field,
        fld.fieldStarter( reg ),
        func( val interface{}, path objpath.PathNode ) error {
            fld.Assign( fsb.Value, val )
            return nil
        },
    )
}

func CheckedFunctionsFieldSetBuilder(
    reg *Registry,
    val interface{},
    setters ...*CheckedFieldSetter ) mgRct.FieldSetBuilder {

    res := mgRct.NewFunctionsFieldSetBuilder()
    res.Value = val
    for _, s := range setters { AddCheckedField( res, reg, s ) }
    return res
}

func checkedStructFunc( 
    reg *Registry, 
    fact func() interface{}, 
    setters []*CheckedFieldSetter ) mgRct.StructStartFunc {

    return func( _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        return CheckedFunctionsFieldSetBuilder( reg, fact(), setters... ), nil
    }
}

func CheckedStructFactory( 
    reg *Registry, 
    fact func() interface{},
    setters ...*CheckedFieldSetter ) mgRct.BuilderFactory {

    validateSetters( setters )
    res := mgRct.NewFunctionsBuilderFactory()
    res.StructFunc = checkedStructFunc( reg, fact, setters )
    return res
}
