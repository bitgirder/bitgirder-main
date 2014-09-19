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

    res := NewFunctionsFieldSetBuilder()
    res.Value = val
    for _, s := range setters { AddCheckedField( res, reg, s ) }
    return res
}

type InstanceFactoryFunc func() interface{}

func CheckedStructFunc( 
    reg *Registry, 
    fact InstanceFactoryFunc,
    setters []*CheckedFieldSetter ) mgRct.StructStartFunc {

    validateSetters( setters )
    return func( _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {
        return CheckedFunctionsFieldSetBuilder( reg, fact(), setters... ), nil
    }
}

func CheckedStructFactory( 
    reg *Registry, 
    fact InstanceFactoryFunc,
    setters ...*CheckedFieldSetter ) mgRct.BuilderFactory {

    res := NewFunctionsBuilderFactory()
    res.StructFunc = CheckedStructFunc( reg, fact, setters )
    return res
}

type ListElementFactoryFunc func( reg *Registry ) mgRct.BuilderFactory

func ListElementFactoryFuncForType( 
    typ mg.TypeReference ) ListElementFactoryFunc {

    return func( reg *Registry ) mgRct.BuilderFactory {
        if res, ok := reg.BuilderFactoryForType( typ ); ok { return res }
        return nil
    }
}

type ListElementAddFunc func( l, val interface{} ) interface{}

func CheckedListBuilder(
    reg *Registry,
    listFact InstanceFactoryFunc,
    nextFact ListElementFactoryFunc,
    addElt ListElementAddFunc ) mgRct.ListBuilder {

    res := NewFunctionsListBuilder()
    res.Value = listFact()
    res.NextFunc = func() mgRct.BuilderFactory { return nextFact( reg ) }
    res.AddFunc = func( val interface{}, path objpath.PathNode ) error {
        res.Value = addElt( res.Value, val )
        return nil
    }
    return res
}

func CheckedListFactory(
    reg *Registry,
    listFact InstanceFactoryFunc,
    nextFact ListElementFactoryFunc,
    addElt ListElementAddFunc ) mgRct.BuilderFactory {

    res := NewFunctionsBuilderFactory()
    res.ListFunc = func( 
        _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {

        return CheckedListBuilder( reg, listFact, nextFact, addElt ), nil
    }
    return res
}

func CheckedListFieldStarter(
    listFact InstanceFactoryFunc,
    nextFact ListElementFactoryFunc,
    addElt ListElementAddFunc ) CheckedFieldStartFunction {

    return func( reg *Registry ) mgRct.BuilderFactory {
        res := NewFunctionsBuilderFactory()
        res.ListFunc = func( 
            _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {
    
            return CheckedListBuilder( reg, listFact, nextFact, addElt ), nil
        }
        return res
    }
}
