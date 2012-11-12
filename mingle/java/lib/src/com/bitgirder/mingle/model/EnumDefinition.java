package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Map;

public
final
class EnumDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< MingleIdentifier > values;
    private final Map< MingleIdentifier, MingleEnum > byName;

    // Important to note that this is called after super (getName() is valid)
    // and after assigning this.values
    private
    Map< MingleIdentifier, MingleEnum >
    buildByName()
    {
        Map< MingleIdentifier, MingleEnum > res = Lang.newMap( values.size() );
        
        AtomicTypeReference typ = AtomicTypeReference.create( getName() );

        for ( MingleIdentifier v : values )
        {
            res.put( v, MingleEnum.create( typ, v ) );
        }

        return Lang.unmodifiableMap( res );
    }

    private
    EnumDefinition( Builder b )
    {
        super( b );

        this.values = inputs.notNull( b.values, "values" );
        this.byName = buildByName();
    }

    public List< MingleIdentifier > getNames() { return values; }

    public
    MingleEnum
    getEnumValue( MingleIdentifier id )
    {
        return byName.get( inputs.notNull( id, "id" ) );
    }

    final
    static
    class Builder
    extends TypeDefinition.Builder< EnumDefinition, Builder >
    {
        private List< MingleIdentifier > values;

        public
        Builder
        setValues( List< MingleIdentifier > values )
        {
            this.values = Lang.unmodifiableCopy( values, "values" );
            inputs.isFalse( values.isEmpty(), "need at least one value" );

            return this;
        }

        public EnumDefinition build() { return new EnumDefinition( this ); }
    }

    public
    static
    EnumDefinition
    create( QualifiedTypeName name,
            List< MingleIdentifier > values )
    {
        return new Builder().setName( name ).setValues( values ).build();
    }
}
