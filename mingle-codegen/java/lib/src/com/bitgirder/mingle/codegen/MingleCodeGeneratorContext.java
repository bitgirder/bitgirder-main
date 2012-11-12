package com.bitgirder.mingle.codegen;

import com.bitgirder.log.CodeLogger;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.mingle.codec.MingleCodecFactory;

import com.bitgirder.mingle.runtime.MingleRuntime;

import java.util.List;

public
interface MingleCodeGeneratorContext
{
    public
    MingleRuntime
    runtime();

    public
    MingleCodecFactory
    codecFactory();

    public
    List< QualifiedTypeName >
    getTargets();

    public
    List< MingleStruct >
    controlObjects();

    public
    CodeLogger
    log();

    public
    void
    writeAsync( CharSequence relName,
                CharSequence sourceText )
        throws Exception;

    public
    void
    complete();
}
