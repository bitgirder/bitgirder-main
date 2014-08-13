package builtin

import (
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/bind"
    "mingle/types"
)

func ( t *BuiltinTypeTest ) createBindReactor( 
    c *mgRct.ReactorTestCall ) *mgRct.BuildReactor {

    reg := bind.RegistryForDomain( bind.DomainDefault )
    if bf, ok := reg.BuilderFactoryForType( t.Type ); ok {
        return bind.NewBuildReactor( bf )
    }
    c.Fatalf( "no binder for type: %s", t.Type )
    panic( libError( "unreachable" ) )
}

func ( t *BuiltinTypeTest ) Call( c *mgRct.ReactorTestCall ) {
    c.Logf( "expcting %s as type: %s", mg.QuoteValue( t.In ), t.Type )
    br := t.createBindReactor( c )
    cr := types.NewCastReactor( t.Type, types.BuiltinTypes() )
    pip := mgRct.InitReactorPipeline( cr, mgRct.NewDebugReactor( c ), br )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        c.Equal( t.Expect, br.GetValue() )
    } else { c.EqualErrors( t.Err, err ) }
}
