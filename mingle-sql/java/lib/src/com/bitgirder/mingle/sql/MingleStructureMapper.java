package com.bitgirder.mingle.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;

import com.bitgirder.sql.AbstractSqlParameterMapper;
import com.bitgirder.sql.SqlParameterGroupDescriptor;
import com.bitgirder.sql.SqlParameterDescriptor;
import com.bitgirder.sql.Sql;

import java.util.Map;
import java.util.Set;

import java.sql.PreparedStatement;

public
final
class MingleStructureMapper< M >
extends AbstractSqlParameterMapper< MingleStructure, M >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< String, ParameterHandler< ? super M > > handlers;
 
    private 
    MingleStructureMapper( Builder< M > b ) 
    { 
        super( inputs.notNull( b.params, "params" ) ); 

        this.handlers = Lang.unmodifiableCopy( b.handlers );
        assertMappingSets();
    }

    // mutates s1
    private
    void
    assertContains( Set< String > s1,
                    Set< String > s2,
                    String s1EltName,
                    String s2EltName )
    {
        s1.removeAll( s2 );

        state.isTrue( 
            s1.isEmpty(), 
            "One or more", s1EltName, "are not amongst the", s2EltName,
            "for this instance:", s1
        );
    } 

    private
    void
    assertMappingSets()
    {
        // make a copy since we'll mutate it
        Set< String > hndlrSet = Lang.newSet( handlers.keySet() );

        Set< String > paramSet = Lang.newSet();

        for ( SqlParameterDescriptor p : parameters().getParameters() )
        {
            paramSet.add( p.getName() );
        }
        
        assertContains( 
            hndlrSet, paramSet, "handlers", "parameter descriptors" );

        assertContains( 
            paramSet, handlers.keySet(), "parameter descriptors", "handlers" );
    }

    protected
    boolean
    setParameter( MingleStructure ms,
                  M mpObj,
                  SqlParameterDescriptor param,
                  PreparedStatement ps,
                  int indx )
        throws Exception
    {
        ParameterHandler< ? super M > h = 
            state.get( handlers, param.getName(), "handlers" );

        return h.setParameter( ms, mpObj, param, ps, indx );
    }

    public
    static
    interface ParameterHandler< M >
    {
        public
        boolean
        setParameter( MingleStructure ms,
                      M mpObj,
                      SqlParameterDescriptor param,
                      PreparedStatement ps,
                      int indx )
            throws Exception;
    }

    public
    static
    abstract
    class AbstractParameterHandler< M >
    implements ParameterHandler< M >
    {
        protected
        final
        MingleSymbolMapAccessor
        mapAccessor( MingleStructure ms )
        {
            return MingleSymbolMapAccessor.create( inputs.notNull( ms, "ms" ) );
        }
    }

    // Although the return type V of getValue() is opaque to the implementation,
    // we make it a typed parameter to help implementors and the compiler have
    // greater assurance that the expected value is being returned 
    public
    static
    abstract
    class AbstractValueMapper< V, M >
    extends AbstractParameterHandler< M >
    {
        protected
        V
        getValue()
            throws Exception
        {
            throw
                state.createFail(
                    "Some form of getValue() must be overridden" );
        }

        protected
        V
        getValue( MingleStructure ms )
            throws Exception
        {
            return getValue();
        }

        protected
        V
        getValue( MingleStructure ms,
                  M mprObj )
            throws Exception
        {
            return getValue( ms );
        }

        protected
        V
        getValue( MingleStructure ms,
                  M mprObj,
                  SqlParameterDescriptor param )
            throws Exception
        {
            return getValue( ms, mprObj );
        }

        public
        final
        boolean
        setParameter( MingleStructure ms,
                      M mprObj,
                      SqlParameterDescriptor param,
                      PreparedStatement ps,
                      int indx )
            throws Exception
        {
            V val = getValue( ms, mprObj, param );
            
            if ( val instanceof MingleValue )
            {
                MingleSql.
                    setValue( (MingleValue) val, param.getSqlType(), ps, indx );
            }
            else Sql.setValue( val, param.getSqlType(), ps, indx );

            return true;
        }
    }

    private
    final
    static
    class ConstantHandler
    extends AbstractValueMapper< Object, Object >
    {
        private final Object val;

        private ConstantHandler( Object val ) { this.val = val; }

        @Override protected Object getValue() { return val; }
    }

    private
    final
    static
    class StructureFieldHandler
    extends AbstractValueMapper< MingleValue, Object >
    {
        private final MingleIdentifier fld;

        private 
        StructureFieldHandler( MingleIdentifier fld ) 
        { 
            this.fld = fld;
        }

        @Override
        protected
        MingleValue
        getValue( MingleStructure ms )
        {
            return ms.getFields().get( fld );
        }
    }

    public
    final
    static
    class Builder< M >
    {
        private final Map< String, ParameterHandler< ? super M > > handlers = 
            Lang.newMap();

        private SqlParameterGroupDescriptor params;

        public
        Builder< M >
        setParameters( SqlParameterGroupDescriptor params )
        {
            this.params = inputs.notNull( params, "params" );
            return this;
        }

        public
        Builder< M >
        map( String prmNm,
             ParameterHandler< ? super M > m )
        {
            Lang.putUnique( 
                handlers, 
                inputs.notNull( prmNm, "prmNm" ),
                inputs.notNull( m, "m" )
            );

            return this;
        }

        // val may be null
        public
        Builder< M >
        mapConstant( String prmNm,
                     Object val )
        {
            return map( prmNm, new ConstantHandler( val ) );
        }

        public
        Builder< M >
        mapField( String prmNm,
                  MingleIdentifier fld )
        {
            inputs.notNull( fld, "fld" );
            return map( prmNm, new StructureFieldHandler( fld ) );
        }

        public
        Builder< M >
        mapField( String prmNm,
                  CharSequence fldStr )
        {
            inputs.notNull( fldStr, "fldStr" );

            MingleIdentifier fld = MingleIdentifier.create( fldStr );

            return mapField( prmNm, fld );
        }

        public
        Builder< M >
        mapField( MingleIdentifier fld )
        {
            inputs.notNull( fld, "fld" );

            String colNm =
                MingleModels.
                    format( fld, MingleIdentifierFormat.LC_UNDERSCORE ).
                    toString();

            return mapField( colNm, fld );
        }

        public
        Builder< M >
        mapField( CharSequence prmNm )
        {
            inputs.notNull( prmNm, "prmNm" );
            return mapField( MingleIdentifier.create( prmNm ) );
        }

        public
        MingleStructureMapper< M >
        build()
        {
            return new MingleStructureMapper< M >( this );
        }
    }
}
