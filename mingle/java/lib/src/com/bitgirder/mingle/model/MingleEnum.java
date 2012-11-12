package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

public
final
class MingleEnum
extends AbstractTypedMingleValue
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleIdentifier value;

    private
    MingleEnum( Builder b )
    {
        super( b );
        this.value = inputs.notNull( b.value, "value" );
    }

    public MingleIdentifier getValue() { return value; }

    public
    static
    MingleEnum
    parse( CharSequence cs )
        throws SyntaxException
    {
        return MingleParsers.parseEnumLiteral( inputs.notNull( cs, "cs" ) );
    }

    public
    static
    MingleEnum
    create( CharSequence cs )
    {
        return MingleParsers.createEnumLiteral( inputs.notNull( cs, "cs" ) );
    }

    public
    static
    MingleEnum
    create( AtomicTypeReference ref,
            MingleIdentifier value )
    {
        inputs.notNull( ref, "ref" );
        inputs.notNull( value, "value" );

        return 
            new Builder().
                setType( ref ).
                setValue( value ).
                build();
    }

    public
    final
    static
    class Builder
    extends MingleTypedValueBuilder< Builder, MingleEnum >
    {
        private MingleIdentifier value;

        public
        Builder
        setValue( MingleIdentifier value )
        {
            this.value = inputs.notNull( value, "value" );
            return this;
        }

        public
        Builder
        setValue( CharSequence value )
        {
            inputs.notNull( value, "value" );
            return setValue( MingleIdentifier.create( value ) );
        }

        public MingleEnum build() { return new MingleEnum( this ); }
    }
}
