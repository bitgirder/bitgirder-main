package types

import (
    mg "mingle"
    "log"
)

var idAuthentication = idUnsafe( "authentication" )

// returns the *PrototypeDefinition matching secQn, but does not check that it
// is otherwise valid as a security prototype
func expectProtoDef( 
    qn *mg.QualifiedTypeName, dm DefinitionGetter ) *PrototypeDefinition {
 
    def := MustGetDefinition( qn, dm )
    if protDef, ok := def.( *PrototypeDefinition ); ok { return protDef }
    panic( libErrorf( "not a prototype: %s", qn ) )
}

func MustAuthTypeOf( 
    secQn *mg.QualifiedTypeName, dm DefinitionGetter ) mg.TypeReference {

    protDef := expectProtoDef( secQn, dm )
    flds := protDef.Signature.GetFields()
    if fd := flds.Get( idAuthentication ); fd != nil { return fd.Type }
    panic( libErrorf( "no auth for security: %s", secQn ) )
}

func canAssignToStruct( 
    targ *StructDefinition, def Definition, dm DefinitionGetter ) bool {

    sd, ok := def.( *StructDefinition )
    if ! ok { return false }
    return targ.GetName().Equals( sd.GetName() ) 
}

func canAssignToSchema(
    targ *SchemaDefinition, def Definition, dm DefinitionGetter ) bool {

    if sd, ok := def.( *StructDefinition ); ok {
        log.Printf( "checking whether %s satisfies schema %s", 
            sd.Name, targ.Name )
        return sd.SatisfiesSchema( targ )
    }
    return false
}

func canAssignType( t1, t2 *mg.QualifiedTypeName, dm DefinitionGetter ) bool {
    d1, d2 := MustGetDefinition( t1, dm ), MustGetDefinition( t2, dm )
    switch v1 := d1.( type ) {
    case *StructDefinition: return canAssignToStruct( v1, d2, dm )
    case *SchemaDefinition: return canAssignToSchema( v1, d2, dm )
    }
    return false
}

func canThrowErrorOfType( 
    qn *mg.QualifiedTypeName, 
    sig *CallSignature,
    dm DefinitionGetter ) ( mg.TypeReference, bool ) {

    for _, typ := range sig.Throws {
        if canAssignType( mg.TypeNameIn( typ ), qn, dm ) { return typ, true }
    }
    return nil, false
}
