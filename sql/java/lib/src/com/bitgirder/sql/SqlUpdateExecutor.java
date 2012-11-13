package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.sql.PreparedStatement;
import java.sql.Connection;

import java.util.Collection;

import javax.sql.DataSource;

public
final
class SqlUpdateExecutor< I, M >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static EventHandler DEFAULT_EVENT_HANDLER =  
        new DefaultEventHandler();

    private final DataSource ds;
    private final Collection< ? extends I > batch;

    // maybe null
    private final MapperObjectGenerator< ? super I, ? extends M > mapperObjGen; 

    private final CharSequence sql;
    private final SqlParameterMapper< ? super I, ? super M > mapper;
    private final EventHandler eh;

    private boolean started;

    private
    SqlUpdateExecutor( Builder< I, M > b )
    {
        this.ds = inputs.notNull( b.ds, "ds" );
        this.sql = inputs.notNull( b.sql, "sql" );
        this.mapper = inputs.notNull( b.mapper, "mapper" );
        this.eh = b.eh;
 
        this.batch = inputs.notNull( b.batch, "batch" );
        state.isFalse( batch.isEmpty(), "Batch is empty" );

        this.mapperObjGen = b.mapperObjGen;
    }

    private
    final
    class UpdateOp
    implements ConnectionUser< Void >
    {
        private
        M
        getMapperObject( Connection conn )
            throws Exception
        {
            if ( mapperObjGen == null ) return null;
            else return mapperObjGen.generateMapperObject( batch, conn );
        }

        private
        int
        setParameters( M mapperObj,
                       PreparedStatement ps )
            throws Exception
        {
            int res = 0;

            for ( I obj : batch )
            {
                ps.clearParameters();

                if ( mapper.setParameters( obj, mapperObj, ps, 0 ) )
                {
                    ++res;
                    ps.addBatch();
                }
            }

            return res;
        }

        public
        Void
        useConnection( Connection conn )
            throws Exception
        {
            PreparedStatement ps = conn.prepareStatement( sql.toString() );
            try
            {
                M mapperObj = getMapperObject( conn );
                if ( setParameters( mapperObj, ps ) > 0 ) ps.executeBatch();
                eh.updateComplete( SqlUpdateExecutor.this ); 
            }
            finally { ps.close(); }

            return null;
        }
    }

    public
    void
    start()
        throws Exception
    {
        state.isFalse( started, "start() already called" );
        started = true;

        Sql.useConnection( ds, new UpdateOp() );
    }

    public
    static
    interface EventHandler
    {
        public
        void
        updateComplete( SqlUpdateExecutor e )
            throws Exception;
        
    }

    public
    static
    class DefaultEventHandler
    implements EventHandler
    {
        public void updateComplete() throws Exception {}

        public 
        void 
        updateComplete( SqlUpdateExecutor e ) 
            throws Exception
        {
            updateComplete();
        }
    }

    public
    static
    interface MapperObjectGenerator< I, M >
    {
        public
        M
        generateMapperObject( Collection< ? extends I > batch,
                              Connection conn )
            throws Exception;
    }

    public
    final
    static
    class Builder< I, M >
    {
        private DataSource ds;
        private Collection< ? extends I > batch;
        private MapperObjectGenerator< ? super I, ? extends M > mapperObjGen;
        private CharSequence sql;
        private SqlParameterMapper< ? super I, ? super M > mapper;
        private EventHandler eh = DEFAULT_EVENT_HANDLER;

        public
        Builder< I, M >
        setDataSource( DataSource ds )
        {
            this.ds = inputs.notNull( ds, "ds" );
            return this;
        }

        public
        Builder< I, M >
        setBatch( Collection< ? extends I > batch )
        {
            this.batch = inputs.notNull( batch, "batch" );
            return this;
        }

        public
        Builder< I, M >
        setMapperObjectGenerator( 
            MapperObjectGenerator< ? super I, ? extends M > mapperObjGen )
        {
            this.mapperObjGen = inputs.notNull( mapperObjGen, "mapperObjGen" );
            return this;
        }

        public
        Builder< I, M >
        setSql( CharSequence sql )
        {
            this.sql = inputs.notNull( sql, "sql" );
            return this;
        }

        public
        Builder< I, M >
        setMapper( SqlParameterMapper< ? super I, ? super M > mapper )
        {
            this.mapper = inputs.notNull( mapper, "mapper" );
            return this;
        }

        public
        Builder< I, M >
        setEventHandler( EventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return this;
        }

        public
        Builder< I, M >
        setOnComplete( final Runnable r )
        {
            inputs.notNull( r, "r" );

            return setEventHandler(
                new DefaultEventHandler() {
                    public void updateComplete() { r.run(); }
                }
            );
        }

        public
        SqlUpdateExecutor< I, M >
        build()
        {
            return new SqlUpdateExecutor< I, M >( this ); 
        }
    }
}
