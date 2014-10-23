package types

import (
    mg "mingle"
//    "log"
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
