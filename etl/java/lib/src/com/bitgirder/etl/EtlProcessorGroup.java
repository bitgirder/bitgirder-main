package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.management.ProcessFactory;
import com.bitgirder.process.management.ProcessControl;

import com.bitgirder.mingle.model.MingleIdentifiedName;

import java.util.Map;

public
final
class EtlProcessorGroup
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< MingleIdentifiedName, ProcessFactory< ? > > procs;
    private final ProcessControl< ? > stateMgr;
    private final ProcessControl< ? > lister;
    private final ProcessControl< ? > recRdr;

    private
    EtlProcessorGroup( Builder b )
    {
        this.procs = Lang.unmodifiableCopy( b.procs );

        this.stateMgr = inputs.notNull( b.stateMgr, "stateMgr" );
        this.lister = inputs.notNull( b.lister, "lister" );
        this.recRdr = inputs.notNull( b.recRdr, "recRdr" );
    }

    Map< MingleIdentifiedName, ProcessFactory< ? > >
    getProcessors()
    {
        return procs;
    }

    ProcessControl< ? > getStateManager() { return stateMgr; }
    ProcessControl< ? > getLister() { return lister; }
    ProcessControl< ? > getRecordReader() { return recRdr; }

    public
    final
    static
    class Builder
    {
        private final Map< MingleIdentifiedName, ProcessFactory< ? > > procs =
            Lang.newMap();

        private ProcessControl< ? > stateMgr;
        private ProcessControl< ? > lister;
        private ProcessControl< ? > recRdr;

        // processes generated are assumed to be instances of
        // AbstractEtlProcessor or some other process which handles the
        // lifecycle and requests expected of an etl processor
        public
        Builder
        addProcessor( MingleIdentifiedName nm,
                      ProcessFactory< ? > fact )
        {
            Lang.putUnique(
                procs,
                inputs.notNull( nm, "nm" ),
                inputs.notNull( fact, "fact" )
            );

            return this;
        }

        public
        Builder
        setStateManager( ProcessControl< ? > stateMgr )
        {
            this.stateMgr = inputs.notNull( stateMgr, "stateMgr" );
            return this;
        }

        public
        Builder
        setSourceLister( ProcessControl< ? > lister )
        {
            this.lister = inputs.notNull( lister, "lister" );
            return this;
        }

        public
        Builder
        setRecordReader( ProcessControl< ? > recRdr )
        {
            this.recRdr = inputs.notNull( recRdr, "recRdr" );
            return this;
        }

        public
        EtlProcessorGroup
        build()
        {
            return new EtlProcessorGroup( this );
        }
    }
}
