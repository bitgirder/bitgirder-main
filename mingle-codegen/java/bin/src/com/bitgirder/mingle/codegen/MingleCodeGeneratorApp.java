package com.bitgirder.mingle.codegen;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.application.ApplicationProcess;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;

import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.TypeDefinitionLookup;
import com.bitgirder.mingle.model.TypeDefinitionCollection;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecFactories;

import com.bitgirder.mingle.runtime.MingleRuntime;
import com.bitgirder.mingle.runtime.MingleRuntimes;

import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.DirWrapper;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoProcessor;

import java.nio.ByteBuffer;

import java.util.List;

final
class MingleCodeGeneratorApp
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final FileWrapper input;
    private final String language;
    private final DirWrapper outDir;
    private final List< FileWrapper > ctlFiles;

    private final IoProcessor ioProc = IoProcessor.create( 1 );

    private
    MingleCodeGeneratorApp( Configurator c )
    {
        super( c );

        this.input = inputs.notNull( c.input, "input" );
        this.language = inputs.notNull( c.language, "language" );
        this.outDir = inputs.notNull( c.outDir, "outDir" );
        this.ctlFiles = c.ctlFiles;
    }

    private
    static
    class RuntimeLoad
    {
        private MingleRuntime rt;
        private List< QualifiedTypeName > toGen = Lang.newList();
    }

    private
    RuntimeLoad
    loadRuntime( MingleCodecFactory codecFact )
        throws Exception
    {
        RuntimeLoad res = new RuntimeLoad();
        
        TypeDefinitionLookup.Builder lkBld = new TypeDefinitionLookup.Builder();

        lkBld.addTypes( MingleRuntimes.loadDefault( codecFact ).getTypes() );

        TypeDefinitionCollection coll =
            MingleRuntimes.asTypeDefinitionCollection(
                IoUtils.toByteBuffer( input.openReadStream(), true ),
                codecFact
            );

        lkBld.addTypes( coll );

        for ( TypeDefinition td : coll.getTypes() ) 
        {
            res.toGen.add( td.getName() );
        }
        
        res.rt = new MingleRuntime.Builder().setTypes( lkBld.build() ).build();

        return res;
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        if ( ! hasChildren() ) exit();
    }

    private void done() { ioProc.stop(); }

    private
    MingleStruct
    loadControlObject( FileWrapper ctlFile,
                       MingleCodecFactory codecFact )
        throws Exception
    {
        ByteBuffer bb = IoUtils.toByteBuffer( ctlFile.openReadStream(), true );

        MingleCodec codec = MingleCodecs.detectCodec( codecFact, bb );
        return MingleCodecs.fromByteBuffer( codec, bb, MingleStruct.class );
    }

    private
    List< MingleStruct >
    loadControlObjects( MingleCodecFactory codecFact )
        throws Exception
    {
        List< MingleStruct > res = Lang.newList();

        for ( FileWrapper ctlFile : ctlFiles ) 
        {
            res.add( loadControlObject( ctlFile, codecFact ) );
        }

        return res;
    }

    private
    MingleCodeGeneration
    buildCodeGeneration( MingleCodecFactory codecFact,
                         RuntimeLoad rt )
        throws Exception
    {
        return
            new MingleCodeGeneration.Builder().
                setActivityContext( getActivityContext() ).
                setRuntime( rt.rt ).
                setCodecFactory( codecFact ).
                setTargets( rt.toGen ).
                setOutDir( outDir ).
                setLanguage( language ).
                setIoProcessor( ioProc ).
                setControlObjects( loadControlObjects( codecFact ) ).
                setOnComplete(
                    new AbstractTask() { 
                        protected void runImpl() { done(); } 
                    } 
                ).
                build();
    }

    protected
    void
    startImpl()
        throws Exception
    {
        spawn( ioProc );

        MingleCodecFactory codecFact = MingleCodecFactories.loadDefault();

        RuntimeLoad rt = loadRuntime( codecFact );

        buildCodeGeneration( codecFact, rt ).start();
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private FileWrapper input;
        private String language;
        private DirWrapper outDir;
        private final List< FileWrapper > ctlFiles = Lang.newList();

        @Argument
        private
        void
        setInput( String input )
        {
            this.input = new FileWrapper( input ).assertExists();
        }

        @Argument
        private
        void
        setLanguage( String language )
        {
            this.language = language;
        }

        @Argument
        private
        void
        setOutDir( String outDir )
        {
            this.outDir = new DirWrapper( outDir ).assertExists();
        }

        @Argument
        private
        void
        setControlObject( String ctlFile )
        {
            ctlFiles.add( new FileWrapper( ctlFile ).assertExists() );
        }
    }
}
