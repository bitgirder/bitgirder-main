package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;
import com.bitgirder.log.CodeLogger;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleValidationException;

import java.util.Map;

// mutable methods are not threadsafe
public
final
class MingleServiceCallContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifier ID_AUTHENTICATION =
        MingleIdentifier.create( "authentication" );

    private final static MingleIdentifier ID_PARAMS =
        MingleIdentifier.create( "parameters" );

    private final MingleServiceRequest req;

    private final ObjectPath< MingleIdentifier > pathRoot =
        ObjectPath.newRoot();

    // lazily initialized
    private Map< Object, Object > attachments;

    MingleServiceCallContext( MingleServiceRequest req ) 
    { 
        this.req = inputs.notNull( req, "req" );
    }

//    public CodeLogger log() { return CodeLoggers.getDefaultLogger(); }

    public MingleServiceRequest getRequest() { return req; }

    public
    Map< Object, Object >
    attachments()
    {
        return attachments == null ? attachments = Lang.newMap() : attachments;
    }

    public 
    MingleSymbolMapAccessor 
    getParameters() 
    { 
        return 
            MingleSymbolMapAccessor.
                create( req.getParameters(), getParametersPath() );
    }

    public MingleSymbolMap getRawParameters() { return req.getParameters(); }

    public
    ObjectPath< MingleIdentifier >
    getParametersPath() 
    { 
        return pathRoot.descend( ID_PARAMS ); 
    }

    public
    ObjectPath< MingleIdentifier >
    getParameterPath( MingleIdentifier key )
    {
        return getParametersPath().descend( inputs.notNull( key, "key" ) );
    }

    public MingleValue getAuthentication() { return req.getAuthentication(); }

    public
    ObjectPath< MingleIdentifier >
    getAuthenticationPath()
    {
        return pathRoot.descend( ID_AUTHENTICATION );
    }

    boolean
    isInboundValidationException( MingleValidationException mve )
    {
        state.notNull( mve, "mve" );
        
        return ObjectPaths.rootOf( mve.getLocation() ) == pathRoot;
    }

    public
    static
    MingleServiceCallContext
    create( MingleServiceRequest req )
    {
        return new MingleServiceCallContext( req );
    }
}
