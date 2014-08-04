package service

import (
    mg "mingle"
)

var NsService *mg.Namespace
var QnameRequest *mg.QualifiedTypeName

type RequestContext struct {
    Namespace *mg.Namespace
    Service *mg.Identifier
    Operation *mg.Identifier
}

func initNames() {
    mkId := func( parts ...string ) *mg.Identifier {
        return mg.NewIdentifierUnsafe( parts )
    }
    declNm := mg.NewDeclaredTypeNameUnsafe
    NsService = &mg.Namespace{
        Parts: []*mg.Identifier{ mkId( "mingle" ), mkId( "service" ) },
        Version: mkId( "v1" ),
    }
    QnameRequest = &mg.QualifiedTypeName{
        Namespace: NsService,
        Name: declNm( "Request" ),
    }
}
