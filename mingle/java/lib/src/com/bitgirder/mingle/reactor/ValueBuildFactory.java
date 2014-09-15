package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleSymbolMap;
import com.bitgirder.mingle.MingleStruct;
import com.bitgirder.mingle.MingleList;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

import com.bitgirder.lang.path.ObjectPath;

public
final
class ValueBuildFactory
extends AbstractBuildFactory
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    @Override public Object buildValue( MingleValue mv ) { return mv; }

    private
    final
    static
    class FieldSetBuilder
    extends AbstractFieldSetBuilder
    {
        private final MingleSymbolMap.BuilderImpl< ?, ? > b;

        private
        FieldSetBuilder( MingleSymbolMap.BuilderImpl< ?, ? > b )
        {
            this.b = b;
        }

        @Override protected Object produceValue() { return b.build(); }

        @Override
        protected
        BuildReactor.Factory
        startField()
        {
            return new ValueBuildFactory();
        }

        @Override
        protected
        void
        setValue( MingleIdentifier fld,
                  Object val )
        {
            b.set( fld, (MingleValue) val );
        }
    }

    @Override
    protected
    BuildReactor.FieldSetBuilder
    startStruct( QualifiedTypeName typ )
    {
        return new FieldSetBuilder( new MingleStruct.Builder().setType( typ ) );
    }

    @Override
    protected
    BuildReactor.FieldSetBuilder
    startMap()
    {
        return new FieldSetBuilder( new MingleSymbolMap.Builder() );
    }

    private
    final
    static
    class ListBuilder
    extends AbstractListBuilder
    {
        private final MingleList.Builder b = new MingleList.Builder();

        @Override protected Object produceValue() { return b.buildLive(); }

        @Override
        protected 
        void 
        addValue( Object val ) 
        { 
            b.addUnsafe( (MingleValue) val ); 
        }

        public
        BuildReactor.Factory
        nextFactory()
        {
            return new ValueBuildFactory(); 
        }
    }

    @Override
    protected 
    BuildReactor.ListBuilder 
    startList( ListTypeReference lt ) 
    { 
        ListBuilder res = new ListBuilder(); 
        res.b.setType( lt );
        
        return res;
    }
}
