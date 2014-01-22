package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.ArrayList;
import java.util.List;

public
final
class MingleValueReactorPipeline
implements MingleValueReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ArrayList< Element< ? > > elements;

    private
    static
    abstract
    class Element< T >
    implements MingleValueReactor
    {
        T obj;

        MingleValueReactor next;
    }

    private
    static
    class ReactorElement
    extends Element< MingleValueReactor >
    {
        public
        void
        processEvent( MingleValueReactorEvent ev )
            throws Exception
        {
            obj.processEvent( ev );
            next.processEvent( ev );
        }
    }

    private
    MingleValueReactorPipeline( Builder b )
    {
        this.elements = new ArrayList< Element< ? > >( b.elements );
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        inputs.notNull( ev, "ev" );
        elements.get( 0 ).processEvent( ev );
    }

    public
    < V >
    V
    elementOfType( Class< V > cls )
    {
        inputs.notNull( cls, "cls" );

        for ( int i = elements.size() - 1; i >= 0; --i ) {
            Element< ? > elt = elements.get( i );
            if ( cls.isInstance( elt.obj ) ) return cls.cast( elt.obj );
        }

        return null;
    }

    public
    final
    static
    class Builder
    {
        private final List< Element< ? > > elements = Lang.newList();

        public
        Builder
        addReactor( MingleValueReactor rct )
        {
            inputs.notNull( rct, "rct" );

            ReactorElement elt = new ReactorElement();
            elt.obj = rct;
            elements.add( elt );

            return this;
        }

        private
        void
        initElements()
        {
            for ( int i = 0, e = elements.size(); i < e; ++i )
            {
                Element< ? > elt = elements.get( i );
    
                int nextIdx = i + 1;
    
                elt.next = nextIdx == e ?
                    MingleValueReactors.discardReactor() : 
                    elements.get( nextIdx );
            }
        }

        public
        MingleValueReactorPipeline
        build()
        {
            inputs.isFalse( elements.isEmpty(), "pipeline is empty" );

            initElements();

            return new MingleValueReactorPipeline( this );
        }
    }
}
