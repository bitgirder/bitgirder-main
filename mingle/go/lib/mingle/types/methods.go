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
 
    if def, ok := dm.GetOk( qn ); ok {
        if protDef, ok := def.( *PrototypeDefinition ); ok { return protDef }
        panic( libErrorf( "not a prototype: %s", qn ) )
    }
    panic( libErrorf( "no def for qname: %s", qn ) )
}

func expectAuthTypeOf( 
    secQn *mg.QualifiedTypeName, dm *DefinitionMap ) mg.TypeReference {

    protDef := expectProtoDef( secQn, dm )
    flds := protDef.Signature.GetFields()
    if fd := flds.Get( mg.IdAuthentication ); fd != nil { return fd.Type }
    panic( libErrorf( "no auth for security: %s", secQn ) )
}

func canThrowErrorOfType( 
    qn *mg.QualifiedTypeName, sig *CallSignature ) ( mg.TypeReference, bool ) {

    for _, typ := range sig.Throws {
        if mg.TypeNameIn( typ ).Equals( qn ) { return typ, true }
    }
    return nil, false
}
