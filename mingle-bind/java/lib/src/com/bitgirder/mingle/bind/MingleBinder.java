package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPathFormatter;

import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleBuffer;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleFloat;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.PrimitiveDefinition;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.TypeDefinitionLookup;

import java.util.Map;

import java.nio.ByteBuffer;

public
final
class MingleBinder
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleModels.ValueErrorFactory ERR_FACT_IMPL =
        new MingleModels.ValueErrorFactory() 
        {
            public 
            RuntimeException 
            createFail( ObjectPath< String > path,
                        String msg )
            {
                return MingleBindingException.createOutbound( msg, path );
            }
        };

    private final static Map< QualifiedTypeName, MingleBinding > DEFAULTS;

    private final Map< QualifiedTypeName, MingleBinding > byType;
    private final Map< Class< ? >, QualifiedTypeName > byClass;
    private final TypeDefinitionLookup types;

    private 
    MingleBinder( Builder b )
    {
        this.byType = Lang.unmodifiableCopy( b.byType );
        this.byClass = Lang.unmodifiableCopy( b.byClass );
        this.types = inputs.notNull( b.types, "types" );
    }

    public TypeDefinitionLookup getTypes() { return types; }

    // could be made public if needed
    boolean
    hasBindingFor( QualifiedTypeName qn )
    {
        inputs.notNull( qn, "qn" );
        return byType.containsKey( qn );
    }

    QualifiedTypeName
    bindingNameForClass( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );
        return byClass.get( cls );
    }

    private
    < V >
    QualifiedTypeName
    expectQname( AtomicTypeReference typ,
                 ObjectPath< V > path,
                 ObjectPathFormatter< ? super V > fmtr )
    {
        AtomicTypeReference.Name nm = typ.getName();

        if ( nm instanceof QualifiedTypeName ) return (QualifiedTypeName) nm;
        else
        {
            throw MingleBindingException.create(
                    "Not a qualified type: " + typ, path, fmtr );
        }
    }

    private
    < V >
    MingleBinding
    bindingFor( AtomicTypeReference typ,
                ObjectPath< V > path,
                ObjectPathFormatter< ? super V > fmtr )
    {
        QualifiedTypeName qn = expectQname( typ, path, fmtr );

        MingleBinding res = byType.get( qn );

        if ( res == null )
        {
            throw 
                MingleBindingException.create(
                    "No binding for type " + qn, path, fmtr );
        }
        else return res;
    }

    public
    MingleValue
    asMingleValue( AtomicTypeReference typ,
                   Object obj,
                   ObjectPath< String > path )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( obj, "obj" );
        inputs.notNull( path, "path" );

        MingleBinding mb = 
            bindingFor( typ, path, MingleBindingException.OUTBOUND_FORMATTER );

        return mb.asMingleValue( obj, this, path );
    }

    public
    Object
    asJavaValue( AtomicTypeReference typ,
                 MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        MingleBinding mb = 
            bindingFor( typ, path, MingleBindingException.INBOUND_FORMATTER );
        
        return mb.asJavaValue( typ, mv, this, path );
    }

    public
    final
    static
    class Builder
    {
        private final Map< QualifiedTypeName, MingleBinding > byType =
            Lang.newMap();

        private final Map< Class< ? >, QualifiedTypeName > byClass =
            Lang.newMap();
        
        private TypeDefinitionLookup types;

        private
        void
        addOptDefaults()
        {
            for ( Map.Entry< QualifiedTypeName, MingleBinding > e :
                    DEFAULTS.entrySet() )
            {
                if ( ! byType.containsKey( e.getKey() ) )
                {
                    byType.put( e.getKey(), e.getValue() );
                }
            }
        }

        public
        Builder
        addBinding( QualifiedTypeName qn,
                    MingleBinding binding )
        {
            inputs.notNull( qn, "qn" );
            inputs.notNull( binding, "binding" );

            Lang.putUnique( byType, qn, binding );

            return this;
        }

        public
        Builder
        addBinding( QualifiedTypeName qn,
                    MingleBinding binding,
                    Class< ? > cls )
        {
            inputs.notNull( qn, "qn" );
            inputs.notNull( binding, "binding" );
            inputs.notNull( cls, "cls" );

            addBinding( qn, binding );
            Lang.putUnique( byClass, cls, qn );

            return this;
        }

        public
        Builder
        setTypes( TypeDefinitionLookup types )
        {
            this.types = inputs.notNull( types, "types" );
            return this;
        }

        public 
        MingleBinder 
        build() 
        { 
            addOptDefaults();
            return new MingleBinder( this ); 
        }
    }

    private
    static
    abstract
    class PrimBinding< J, M extends MingleValue >
    implements MingleBinding
    {
        private final Class< J > jvCls;
        private final Class< M > mgCls;

        private
        PrimBinding( Class< J > jvCls,
                     Class< M > mgCls )
        {
            this.jvCls = jvCls;
            this.mgCls = mgCls;
        }

        // Typed as J to help enforce that we're returning what we expect
        abstract
        J
        asJavaValue( M mgVal );
                
        public
        Object
        asJavaValue( AtomicTypeReference typ,
                     MingleValue mv,
                     MingleBinder mb,
                     ObjectPath< MingleIdentifier > path )
        {
            M mgVal = 
                mgCls.cast( MingleModels.asMingleInstance( typ, mv, path ) );

            return asJavaValue( mgVal );
        }

        public
        final
        MingleValue
        asMingleValue( Object obj,
                       MingleBinder mb,
                       ObjectPath< String > path )
        {
            return
                MingleModels.
                    asMingleValue( jvCls.cast( obj ), path, ERR_FACT_IMPL );
        }
    }

    static
    {
        Map< QualifiedTypeName, MingleBinding > m = Lang.newMap();

        m.put(
            PrimitiveDefinition.QNAME_VALUE,
            new PrimBinding< MingleValue, MingleValue >(
                MingleValue.class, MingleValue.class )
            {
                MingleValue asJavaValue( MingleValue mv ) { return mv; }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_NULL,
            new PrimBinding< MingleNull, MingleNull >(
                MingleNull.class, MingleNull.class )
            {
                MingleNull asJavaValue( MingleNull n ) { return null; }
            }
        );

        m.put( 
            PrimitiveDefinition.QNAME_STRING,
            new PrimBinding< String, MingleString >( 
                String.class, MingleString.class ) 
            {
                String asJavaValue( MingleString s ) { return s.toString(); }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_INT64,
            new PrimBinding< Long, MingleInt64 >(
                Long.class, MingleInt64.class )
            {
                Long asJavaValue( MingleInt64 i ) { return i.longValue(); }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_INT32,
            new PrimBinding< Integer, MingleInt32 >(
                Integer.class, MingleInt32.class )
            {
                Integer asJavaValue( MingleInt32 i ) { return i.intValue(); }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_DOUBLE,
            new PrimBinding< Double, MingleDouble >(
                Double.class, MingleDouble.class )
            {
                Double asJavaValue( MingleDouble d ) { return d.doubleValue(); }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_FLOAT,
            new PrimBinding< Float, MingleFloat >(
                Float.class, MingleFloat.class )
            {
                Float asJavaValue( MingleFloat d ) { return d.floatValue(); }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_BOOLEAN,
            new PrimBinding< Boolean, MingleBoolean >(
                Boolean.class, MingleBoolean.class )
            {
                Boolean 
                asJavaValue( MingleBoolean b )
                {
                    return Boolean.valueOf( b.booleanValue() );
                }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_BUFFER,
            new PrimBinding< ByteBuffer, MingleBuffer >(
                ByteBuffer.class, MingleBuffer.class )
            {
                ByteBuffer
                asJavaValue( MingleBuffer b )
                {
                    return b.getByteBuffer();
                }
            }
        );

        m.put(
            PrimitiveDefinition.QNAME_TIMESTAMP,
            new PrimBinding< MingleTimestamp, MingleTimestamp >(
                MingleTimestamp.class, MingleTimestamp.class )
            {
                MingleTimestamp asJavaValue( MingleTimestamp t ) { return t; }
            }
        );

        DEFAULTS = Lang.unmodifiableMap( m );
    }
}
