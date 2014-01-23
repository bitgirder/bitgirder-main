package com.bitgirder.pipeline;

public
interface PipelineInitializer< V >
{
    public
    void
    initialize( PipelineInitializationContext< V > ctx )
        throws Exception;
}
