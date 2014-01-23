package com.bitgirder.pipeline;

public
interface PipelineInitializationContext< V >
{
    // the current pipeline; not guaranteed to remain the same across the life
    // of this instance
    public Pipeline< V > pipeline();

    // If elt is also a PipelineInitializer, it is assumed to be of type
    // PipelineInitializer< V > and will have its initialization method called
    // with this instance before elt is added to the end of the pipeline.
    public
    void
    addElement( V elt );
}
