package com.bitgirder.pipeline;

public
interface PipelineInitializer< V >
{
    public
    void
    initialize( PipelineInitializerContext< V > ctx );
}
