package builtin

import (
    mg "mingle"
    "mingle/types"
)

func matchIdPathPart( in types.UnionMatchInput ) ( mg.TypeReference, bool ) {
    if typ, ok := in.Union.MatchType( in.TypeIn ); ok { return typ, ok }
    def, ok := in.Definitions.GetDefinition( mg.QnameIdentifier )
    if ! ok { panic( libErrorf( "no definition for Identifier" ) ) }
    sd := def.( *types.StructDefinition )
    if typ, ok := sd.Constructors.MatchType( in.TypeIn ); ok { return typ, ok }
    if at, ok := in.TypeIn.( *mg.AtomicTypeReference ); ok {
        if mg.IsIntegerTypeName( at.Name() ) { return mg.TypeUint64, true }
    }
    return nil, false
}

func CastBuiltinTypes( cr *types.CastReactor ) {
    cr.SetUnionDefinitionMatcher( mg.QnameIdentifierPathPart, matchIdPathPart )
}
