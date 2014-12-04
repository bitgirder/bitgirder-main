package codegen

import ( 
    mg "mingle"
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
