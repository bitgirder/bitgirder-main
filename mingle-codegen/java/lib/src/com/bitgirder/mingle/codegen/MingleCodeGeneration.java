package com.bitgirder.mingle.codegen;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.log.CodeLogger;

import com.bitgirder.process.ProcessActivity;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.mingle.runtime.MingleRuntime;

import com.bitgirder.mingle.codec.MingleCodecFactory;

import com.bitgirder.io.DirWrapper;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.IoProcessors;
import com.bitgirder.io.Charsets;

import java.nio.ByteBuffer;

import java.nio.channels.FileChannel;

import java.util.List;

public
final
class MingleCodeGeneration
extends ProcessActivity
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleRuntime runtime;
    private final MingleCodecFactory codecFact;
    private final List< QualifiedTypeName > targets;
    private final DirWrapper outDir;
    private final String language;
    private final IoProcessor ioProc;
    private final Runnable onComplete;
    private final List< MingleStruct > ctlObjs;

    private IoProcessor.Client ioCli;
    private boolean started;
    private int waitCount;

    private
    MingleCodeGeneration( Builder b )
    {
        super( b );

        this.runtime = inputs.notNull( b.runtime, "runtime" );
        this.codecFact = inputs.notNull( b.codecFact, "codecFact" );
        this.targets = inputs.notNull( b.targets, "targets" );
        this.outDir = inputs.notNull( b.outDir, "outDir" );
        this.language = inputs.notNull( b.language, "language" );
        this.ioProc = inputs.notNull( b.ioProc, "ioProc" );
        this.onComplete = inputs.notNull( b.onComplete, "onComplete" );
        this.ctlObjs = Lang.unmodifiableCopy( b.ctlObjs );
    }
 
    private MingleCodeGeneration genSelf() { return MingleCodeGeneration.this; }

    private
    void
    completeConditional()
    {
        state.isFalse( waitCount < 0 );

        if ( --waitCount == -1 ) onComplete.run();
    }

    private
    Class< ? extends MingleCodeGenerator >
    getGeneratorClass()
        throws Exception
    {
        String clsNm;
        String pref = "com.bitgirder.mingle.codegen";

        if ( language.equals( "java" ) ) clsNm = pref + ".java.JvCodeGenerator";
        else throw state.createFail( "Unhandled language:", language );

        return Class.forName( clsNm ).asSubclass( MingleCodeGenerator.class );
    }

    private
    MingleCodeGenerator
    createCodeGenerator()
        throws Exception
    {
        return ReflectUtils.newInstance( getGeneratorClass() );
    }

    private
    final
    class DrainHandler
    extends IoProcessors.AbstractFileIoHandler
    {
        private final FileChannel chan;
        private final FileWrapper file;

        private
        DrainHandler( FileChannel chan,
                      FileWrapper file )
        {
            super( genSelf() );
            
            this.chan = chan;
            this.file = file;
        }

        @Override 
        protected 
        void 
        completeIoSucceededImpl() 
            throws Exception
        {
            chan.close();
            code( "Wrote", file );

            completeConditional();
        }
        
        @Override
        protected void completeIoFailed( Throwable th ) throws Exception
        {
            chan.close();
            super.completeIoFailed( th );
        }
    }

    private
    void
    writeSource( CharSequence relName,
                 ByteBuffer src )
        throws Exception
    {
        FileWrapper file = new FileWrapper( outDir, relName );
        file.dirName().mkdirs();
        FileChannel chan = file.openWriteChannel();
        
        code( "Writing", file );

        IoProcessors.drain( 
            ioCli, 
            chan, 
            chan.position(), 
            src, 
            new DrainHandler( chan, file ) 
        );
    }

    private
    final
    class ContextImpl
    implements MingleCodeGeneratorContext
    {
        public CodeLogger log() { return genSelf().log(); }
        public MingleRuntime runtime() { return runtime; }
        public MingleCodecFactory codecFactory() { return codecFact; }
        public List< MingleStruct > controlObjects() { return ctlObjs; }
        public List< QualifiedTypeName > getTargets() { return targets; }

        public
        void
        writeAsync( CharSequence relName,
                    CharSequence sourceText )
            throws Exception
        {
            inputs.notNull( relName, "relName" );
            inputs.notNull( sourceText, "sourceText" );

            state.isTrue( waitCount >= 0, "This context is completed" );
            
            ByteBuffer src = Charsets.UTF_8.asByteBuffer( sourceText );

            ++waitCount;
            writeSource( relName, src );
        }

        public void complete() { completeConditional(); }
    }
 
    public
    void
    start()
        throws Exception
    {
        state.isFalse( started, "start() already called" );
        started = true;

        ioCli = ioProc.createClient( getActivityContext() );
        createCodeGenerator().start( new ContextImpl() );
    }

    final
    static
    class Builder
    extends ProcessActivity.Builder< Builder >
    {
        private MingleRuntime runtime;
        private MingleCodecFactory codecFact;
        private List< QualifiedTypeName > targets;
        private DirWrapper outDir;
        private String language;
        private IoProcessor ioProc;
        private Runnable onComplete;
        private List< MingleStruct > ctlObjs = Lang.emptyList();

        public
        Builder
        setRuntime( MingleRuntime runtime )
        {
            this.runtime = inputs.notNull( runtime, "runtime" );
            return this;
        }

        public
        Builder
        setCodecFactory( MingleCodecFactory codecFact )
        {
            this.codecFact = inputs.notNull( codecFact, "codecFact" );
            return this;
        }

        public
        Builder
        setTargets( List< QualifiedTypeName > targets )
        {
            this.targets = Lang.unmodifiableCopy( targets, "targets" );
            inputs.isFalse( targets.isEmpty(), "targets is empty" );

            return this;
        }


        public
        Builder
        setOutDir( DirWrapper outDir )
        {
            this.outDir = inputs.notNull( outDir, "outDir" );
            return this;
        }

        public
        Builder
        setLanguage( String language )
        {
            this.language = inputs.notNull( language, "language" );
            return this;
        }

        public
        Builder
        setIoProcessor( IoProcessor ioProc )
        {
            this.ioProc = inputs.notNull( ioProc, "ioProc" );
            return this;
        }

        public
        Builder
        setOnComplete( Runnable onComplete )
        {
            this.onComplete = inputs.notNull( onComplete, "onComplete" );
            return this;
        }

        public
        Builder
        setControlObjects( List< MingleStruct > ctlObjs )
        {
            this.ctlObjs = inputs.noneNull( ctlObjs, "ctlObjs" );
            return this;
        }
        
        public
        MingleCodeGeneration
        build()
        {
            return new MingleCodeGeneration( this );
        }
    }
}
