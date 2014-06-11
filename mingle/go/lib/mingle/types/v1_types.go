package types

import (
    mg "mingle"
)

var coreTypesV1 *DefinitionMap

func CoreTypesV1() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( coreTypesV1 )
    return res
}

func asCoreV1Qn( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( mg.CoreNsV1 )
}

func initCoreV1Prims() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &PrimitiveDefinition{}
        pd.Name = primTyp.Name
        coreTypesV1.MustAdd( pd )
    }
}

func initCoreV1ValueTypes() {
    coreTypesV1.MustAdd( &PrimitiveDefinition{ Name: mg.QnameValue } )
}

func initCoreV1StandardError() *StructDefinition {
    ed := NewStructDefinition()
    ed.Name = asCoreV1Qn( "StandardError" )
    coreTypesV1.MustAdd( ed )
    return ed
}

func newV1StandardError( 
    nm string, ns *mg.Namespace, stdErr *StructDefinition ) *StructDefinition {
    res := NewStructDefinition()
    res.Name = mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( ns )
    res.SuperType = stdErr.Name
    return res
}

func newCoreV1StandardError( 
    nm string, stdErr *StructDefinition ) *StructDefinition {
    return newV1StandardError( nm, mg.CoreNsV1, stdErr )
}

func initCoreV1MissingFieldsError( stdErr *StructDefinition ) {
    ed := newCoreV1StandardError( "MissingFieldsError", stdErr )
    coreTypesV1.MustAdd( ed )
}

func initCoreV1UnrecognizedFieldError( stdErr *StructDefinition ) {
    ed := newCoreV1StandardError( "UnrecognizedFieldError", stdErr )
    coreTypesV1.MustAdd( ed )
}

func initCoreV1ValueCastError( stdErr *StructDefinition ) {
    ed := newCoreV1StandardError( "ValueCastError", stdErr )
    coreTypesV1.MustAdd( ed )
}

func initServiceV1EndpointError( stdErr *StructDefinition ) {
    ed := newCoreV1StandardError( "EndpointError", stdErr )
    coreTypesV1.MustAdd( ed )
}

func initCoreV1Exceptions() {
    stdErr := initCoreV1StandardError()
    initCoreV1MissingFieldsError( stdErr )
    initCoreV1UnrecognizedFieldError( stdErr )
    initCoreV1ValueCastError( stdErr )
    initServiceV1EndpointError( stdErr )
}

func init() {
    coreTypesV1 = NewDefinitionMap()
    initCoreV1Prims()
    initCoreV1ValueTypes()
    initCoreV1Exceptions()
}

// package note: not safe to call before completion of package init
func NewV1DefinitionMap() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( coreTypesV1 )
    coreTypesV1.EachDefinition( func( def Definition ) {
        res.setBuiltIn( def.GetName() )
    })
    return res
}
