package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Iterator;

public
final
class EtlRecordSets
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private EtlRecordSets() {}

    private
    final
    static
    class DefaultRecordSet
    implements EtlRecordSet
    {
        private final List< Object > recs;
        private final boolean isFinal;

        private 
        DefaultRecordSet( List< Object > recs,
                          boolean isFinal ) 
        { 
            this.recs = recs; 
            this.isFinal = isFinal;
        }

        public Iterator< Object > iterator() { return recs.iterator(); }
        public int size() { return recs.size(); }
        public boolean isFinal() { return isFinal; }
    }

    // recs may not be null, but may contain null
    public
    static
    EtlRecordSet
    create( Object[] recs,
            boolean isFinal )
    {
        inputs.notNull( recs, "recs" );

        return new DefaultRecordSet( Lang.asList( recs ), isFinal );
    }
}
