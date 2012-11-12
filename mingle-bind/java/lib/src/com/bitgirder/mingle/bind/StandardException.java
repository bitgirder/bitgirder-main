package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.TypeDefinitionLookup;

public
class StandardException
extends BoundException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static QualifiedTypeName QNAME =
        QualifiedTypeName.create( "mingle:core@v1/StandardException" );

    final static AtomicTypeReference MINGLE_TYPE =
        AtomicTypeReference.create( QNAME );

    private final static ObjectPath< String > JV_PATH_MESSAGE =
        ObjectPath.getRoot( "message" );

    public final static MingleIdentifier ID_MESSAGE =
        MingleIdentifier.create( "message" );
    
    private final static MingleTypeReference MSG_TYPE =
        MingleTypeReference.create( "mingle:core@v1/String?" );

    protected
    StandardException( String msg,
                       Throwable cause )
    {
        super( 
            MingleBinders.validateFieldValue( msg, MSG_TYPE, JV_PATH_MESSAGE ), 
            cause 
        );
    }

    public final String message() { return getMessage(); }

    public
    static
    StandardException
    create( String msg )
    {
        return new StandardException( msg, null );
    }

    public
    static
    StandardException
    create( String msg,
            Throwable cause )
    {
        return new StandardException( msg, cause );
    }

    public
    static
    class AbstractBuilder< B extends AbstractBuilder< B > >
    extends BoundException.AbstractBuilder< B >
    {
        protected String message;

        public
        final
        B
        setMessage( String message )
        {
            this.message = message;
            return castThis();
        }
    }

    public
    final
    static
    class Builder
    extends AbstractBuilder< Builder >
    {
        public
        StandardException
        build()
        {
            return StandardException.create( message, _cause );
        }
    }

    static
    void
    addStandardBinding( MingleBinder.Builder b,
                        TypeDefinitionLookup types )
    {
        state.notNull( b, "b" );
        state.notNull( types, "types" );

        new AbstractBindImplementation().initialize( b, types );
    }

    protected
    static
    class AbstractBindImplementation
    extends BoundException.AbstractBindImplementation
    {
        @Override
        public
        void
        initialize( MingleBinder.Builder b,
                    TypeDefinitionLookup types )
        {
            implSetTypeDef( types, QNAME );
            b.addBinding( QNAME, this );
        }

        @Override
        protected
        void
        implSetFields( Object obj,
                       MingleSymbolMapBuilder b,
                       MingleBinder mb,
                       ObjectPath< String > path )
        {
            super.implSetFields( obj, b, mb, path );

            String msg = ( (StandardException) obj ).getMessage();

            if ( msg != null ) b.setString( ID_MESSAGE, msg );
        }

        @Override
        protected
        Object
        implFromMingleStructure( MingleSymbolMap m,
                                 MingleBinder mb,
                                 ObjectPath< MingleIdentifier > path )
        {
            String msg = null;
            
            MingleValue mv = m.get( ID_MESSAGE );

            if ( mv != null )
            {
                MingleString mgStr = (MingleString)
                    MingleModels.asMingleInstance(
                        MingleModels.TYPE_REF_MINGLE_STRING,
                        mv, 
                        path.descend( ID_MESSAGE )
                    );
                
                msg = mgStr.toString();
            }

            return StandardException.create( msg );
        }
    }
}
