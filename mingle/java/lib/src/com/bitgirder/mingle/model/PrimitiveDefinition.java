package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

public
final
class PrimitiveDefinition
extends TypeDefinition
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleNamespace PRIM_NS =
        MingleModels.NS_MINGLE_CORE;

    public final static QualifiedTypeName QNAME_VALUE =
        MingleTypeName.create( "Value" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_NULL =
        MingleTypeName.create( "Null" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_STRING =
        MingleTypeName.create( "String" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_INT64 =
        MingleTypeName.create( "Int64" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_INT32 =
        MingleTypeName.create( "Int32" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_DOUBLE =
        MingleTypeName.create( "Double" ).resolveIn( PRIM_NS );

    public final static QualifiedTypeName QNAME_FLOAT =
        MingleTypeName.create( "Float" ).resolveIn( PRIM_NS );
    
    public final static QualifiedTypeName QNAME_BOOLEAN =
        MingleTypeName.create( "Boolean" ).resolveIn( PRIM_NS );
    
    public final static QualifiedTypeName QNAME_TIMESTAMP =
        MingleTypeName.create( "Timestamp" ).resolveIn( PRIM_NS );
    
    public final static QualifiedTypeName QNAME_BUFFER =
        MingleTypeName.create( "Buffer" ).resolveIn( PRIM_NS );

    private final static Map< QualifiedTypeName, PrimitiveDefinition > PRIMS;

    private final Class< ? extends MingleValue > mgCls;

    private 
    PrimitiveDefinition( Builder b ) 
    { 
        super( b ); 
        this.mgCls = b.mgCls;
    }

    public Class< ? extends MingleValue > getModelClass() { return mgCls; }

    public static MingleNamespace getPrimitiveNamespace() { return PRIM_NS; }

    public
    static
    Map< QualifiedTypeName, PrimitiveDefinition >
    getAll()
    {
        return PRIMS;
    }

    public
    static
    PrimitiveDefinition
    forName( QualifiedTypeName qn )
    {
        inputs.notNull( qn, "qn" );
        return PRIMS.get( qn );
    }

    public
    static
    Class< ? extends MingleValue >
    modelClassFor( QualifiedTypeName qn )
    {
        PrimitiveDefinition def = forName( qn );
        return def == null ? null : def.mgCls;
    }

    private
    final
    static
    class Builder
    extends TypeDefinition.Builder< PrimitiveDefinition, Builder >
    {
        private Class< ? extends MingleValue > mgCls;

        public 
        PrimitiveDefinition
        build()
        {
            return new PrimitiveDefinition( this );
        }
    }

    private
    static
    void
    addPrim( Map< QualifiedTypeName, PrimitiveDefinition > m,
             QualifiedTypeName qn,
             Class< ? extends MingleValue > mgCls )
    {
        Builder b =
            new Builder().
                setName( qn );
        
        b.mgCls = mgCls;
        m.put( qn, b.build() );
    }

    static
    {
        Map< QualifiedTypeName, PrimitiveDefinition > m = Lang.newMap();

        addPrim( m, QNAME_VALUE, MingleValue.class );
        addPrim( m, QNAME_NULL, MingleNull.class );
        addPrim( m, QNAME_STRING, MingleString.class );
        addPrim( m, QNAME_DOUBLE, MingleDouble.class );
        addPrim( m, QNAME_FLOAT, MingleFloat.class );
        addPrim( m, QNAME_INT64, MingleInt64.class );
        addPrim( m, QNAME_INT32, MingleInt32.class );
        addPrim( m, QNAME_BOOLEAN, MingleBoolean.class );
        addPrim( m, QNAME_TIMESTAMP, MingleTimestamp.class );
        addPrim( m, QNAME_BUFFER, MingleBuffer.class );

        PRIMS = Lang.unmodifiableMap( m );
    }
}
