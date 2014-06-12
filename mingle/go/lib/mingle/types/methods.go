package types

import (
    mg "mingle"
)

func collectFieldSets( sd *StructDefinition, dm *DefinitionMap ) []*FieldSet {
    flds := make( []*FieldSet, 0, 2 )
    for {
        flds = append( flds, sd.Fields )
        spr := sd.GetSuperType()
        if spr == nil { break }
        if def, ok := dm.GetOk( spr ); ok {
            if sd, ok = def.( *StructDefinition ); ! ok {
                tmpl := "super type %s of %s is not a struct"
                panic( libErrorf( tmpl, spr, sd.GetName() ) )
            }
        } else {
            tmpl := "can't find super type %s of %s"
            panic( libErrorf( tmpl, spr, sd.GetName() ) )
        }
    }
    return flds
}

// returns the *PrototypeDefinition matching secQn, but does not check that it
// is otherwise valid as a security prototype
func expectProtoDef( 
    qn *mg.QualifiedTypeName, dm *DefinitionMap ) *PrototypeDefinition {
 
    def := dm.MustGet( qn )
    if protDef, ok := def.( *PrototypeDefinition ); ok { return protDef }
    panic( libErrorf( "not a prototype: %s", qn ) )
}

func expectAuthTypeOf( 
    secQn *mg.QualifiedTypeName, dm *DefinitionMap ) mg.TypeReference {

    protDef := expectProtoDef( secQn, dm )
    flds := protDef.Signature.GetFields()
    if fd := flds.Get( mg.IdAuthentication ); fd != nil { return fd.Type }
    panic( libErrorf( "no auth for security: %s", secQn ) )
}

func canAssignToStruct( 
    targ *StructDefinition, def Definition, dm *DefinitionMap ) bool {

    sd, ok := def.( *StructDefinition )
    if ! ok { return false }
    if targ.GetName().Equals( sd.GetName() ) { return true }
    if spr := sd.SuperType; spr != nil {
        return canAssignToStruct( targ, dm.MustGet( spr ), dm )
    }
    return false
}

func canAssignType( t1, t2 *mg.QualifiedTypeName, dm *DefinitionMap ) bool {
    d1, d2 := dm.MustGet( t1 ), dm.MustGet( t2 )
    switch v1 := d1.( type ) {
    case *StructDefinition: return canAssignToStruct( v1, d2, dm )
    }
    return false
}

func canThrowErrorOfType( 
    qn *mg.QualifiedTypeName, 
    sig *CallSignature,
    dm *DefinitionMap ) ( mg.TypeReference, bool ) {

    for _, typ := range sig.Throws {
        if canAssignType( mg.TypeNameIn( typ ), qn, dm ) { return typ, true }
    }
    return nil, false
}
