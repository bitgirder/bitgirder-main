package com.bitgirder.etl.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.etl.AccumulatingEtlProcessor;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleTypeReference;

import java.util.Set;
import java.util.Iterator;
import java.util.AbstractCollection;
import java.util.Collection;

public
abstract
class MingleRecordProcessor
extends AccumulatingEtlProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Set< MingleTypeReference > acceptTypes = Lang.newSet();
    private boolean filterInitDone;

    protected
    MingleRecordProcessor( Builder< ? > b )
    {
        super( inputs.notNull( b, "b" ) );
    }

    protected MingleRecordProcessor() { this( new Builder() ); }

    private
    void
    checkFilterInit( String methName )
    {
        state.isFalse( 
            filterInitDone,
            "Attempt to call", methName, "after filter init completed"
        );
    }

    protected
    final
    void
    acceptType( MingleTypeReference typ )
    {
        checkFilterInit( "acceptType()" );
        inputs.notNull( typ, "typ" );

        acceptTypes.add( typ );
    }

    protected
    final
    void
    acceptType( CharSequence typ )
    {
        inputs.notNull( typ, "typ" );
        acceptType( MingleTypeReference.create( typ ) );
    }

    protected void initFilter() throws Exception {}
    protected void completeInit() throws Exception {}

    // Subclasses which override must make a call to super.init() before doing
    // their own init
    @Override
    protected
    final
    void
    init()
        throws Exception
    {
        initFilter();
        filterInitDone = true;

        completeInit();
    }

    // may be further overridden, possibly calling super.acceptRecord() first in
    // the overridden version to filter on type and then to further filter based
    // on other attributes
    protected
    boolean
    acceptRecord( Object rec )
    {
        if ( acceptTypes.isEmpty() ) return true;
        else return acceptTypes.contains( ( (MingleStruct) rec ).getType() );
    }

    protected
    abstract
    class MingleBatchProcessor
    extends AbstractBatchProcessor
    {
        private
        abstract
        class BatchConverter< V >
        extends AbstractCollection< V >
        {
            abstract V convert( Object obj );

            public 
            Iterator< V >
            iterator()
            {
                return new Iterator< V >() {
                    private final Iterator< ? > it = batch().iterator();
                    public boolean hasNext() { return it.hasNext(); }
                    public void remove() { it.remove(); }
                    public V next() { return convert( it.next() ); }
                };
            }

            public final int size() { return batch().size(); }
        }

        protected
        final
        Collection< MingleSymbolMapAccessor >
        symbolMaps()
        {
            return new BatchConverter< MingleSymbolMapAccessor >() {
                public MingleSymbolMapAccessor convert( Object obj ) {
                    return MingleSymbolMapAccessor.create( (MingleStruct) obj );
                }
            };
        }

        protected
        final
        Collection< MingleStruct >
        structs()
        {
            return new BatchConverter< MingleStruct >() {
                public MingleStruct convert( Object obj ) {
                    return (MingleStruct) obj;
                }
            };
        }
    }

    public
    static
    class Builder< B extends Builder< B > >
    extends AccumulatingEtlProcessor.Builder< B >
    {}
}
