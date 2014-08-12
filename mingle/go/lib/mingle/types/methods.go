package types

import (
    mg "mingle"
    "log"
)

var idAuthentication = idUnsafe( "authentication" )

// returns the *PrototypeDefinition matching qn
func MustPrototypeDefinition( 
    qn *mg.QualifiedTypeName, dm DefinitionGetter ) *PrototypeDefinition {
 
    def := MustGetDefinition( qn, dm )
    if protDef, ok := def.( *PrototypeDefinition ); ok { return protDef }
    panic( libErrorf( "not a prototype: %s", qn ) )
}

func MustAuthenticationType( pd *PrototypeDefinition ) mg.TypeReference {
    flds := pd.Signature.GetFields()
    if fd := flds.Get( idAuthentication ); fd != nil { return fd.Type }
    panic( libErrorf( "no auth for security: %s", pd.Name ) )
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

func CanFailWithError( 
    qn *mg.QualifiedTypeName, 
    typs []mg.TypeReference,
    dm DefinitionGetter ) ( mg.TypeReference, bool ) {

    for _, typ := range typs {
        if canAssignType( mg.TypeNameIn( typ ), qn, dm ) { return typ, true }
    }
    return nil, false
}
