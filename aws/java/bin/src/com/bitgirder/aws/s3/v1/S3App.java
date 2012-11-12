package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleModels;

import com.bitgirder.mingle.application.MingleApplicationProcess;

final
class S3App
extends MingleApplicationProcess< S3AppCfg >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private 
    S3App()
    {
        super(
            new Builder< S3AppCfg >().
                setConfigClass( S3AppCfg.class ).
                setConfigType( "bitgirder:aws:s3@v1/S3AppCfg" )
        );
    }

    protected
    void
    startApp()
    {
        code( "app starting, op:", MingleModels.inspect( config().op() ) );
    }
}
