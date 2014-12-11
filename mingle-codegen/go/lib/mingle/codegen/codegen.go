package codegen

import ( 
    mg "mingle"
    "mingle/bind"
    mgRct "mingle/reactor"
)

type PathMapper interface {
    MapPath( ns *mg.Namespace ) ( []*mg.Identifier, error )
}

type defaultPathMapper int

func ( m defaultPathMapper ) MapPath( 
    ns *mg.Namespace ) ( []*mg.Identifier, error ) {

    res := make( []*mg.Identifier, 0, len( ns.Parts ) + 1 )
    res = append( res, ns.Parts[ 0 ] )
    res = append( res, ns.Version )
    if len( ns.Parts ) > 1 { res = append( res, ns.Parts[ 1 : ]... ) }
    return res, nil
}

var DefaultPathMapper = defaultPathMapper( 1 )

func MustBuilderFactoryForQname( 
    qn *mg.QualifiedTypeName, reg *bind.Registry ) mgRct.BuilderFactory {

    res, ok := reg.BuilderFactoryForName( qn )
    if ! ok { panic( libErrorf( "no builder factory for type: %s", qn ) ) }
    return res
}
