package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Collection;

public
abstract
class AccumulatingEtlProcessor
extends AbstractEtlProcessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static int DEFAULT_MIN_BATCH_SIZE = 1000;

    private final int minBatchSize;

    private final List< Object > batch;
    private Object lastRec;

    protected
    AccumulatingEtlProcessor( Builder< ? > b )
    {
        super( b );

        this.minBatchSize = b.minBatchSize;
        this.batch = Lang.newList( minBatchSize );
    }

    protected boolean acceptRecord( Object rec ) { return true; }

    protected
    abstract
    class AbstractBatchProcessor
    extends AbstractRecordSetProcessor
    {
        protected
        final
        void
        batchDone( Object procState )
        {
            batch.clear();
            respond( procState );
        }
    
        protected final Collection< ? > batch() { return batch; }
    
        protected
        abstract
        void
        startBatch()
            throws Exception;
    
        protected
        final
        void
        startProcess()
            throws Exception
        {
            for ( Object rec : recordSet() )
            {
                if ( acceptRecord( rec ) ) batch.add( rec );
                lastRec = rec;
            }
    
            if ( recordSet().isFinal() )
            {
                if ( batch.isEmpty() ) batchDone( null ); else startBatch();
            }
            else
            {
                if ( batch.size() >= minBatchSize ) startBatch();
                else respond( null );
            }
        }
    }

    public
    static
    class Builder< B extends Builder< B > >
    extends AbstractEtlProcessor.Builder< B >
    {
        private int minBatchSize = DEFAULT_MIN_BATCH_SIZE;

        public
        final
        B
        setMinBatchSize( int minBatchSize )
        {
            this.minBatchSize = 
                inputs.positiveI( minBatchSize, "minBatchSize" );

            return castThis();
        }
    }
}
