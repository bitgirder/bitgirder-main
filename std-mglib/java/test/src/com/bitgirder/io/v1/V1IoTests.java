package com.bitgirder.io.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.DirWrapper;
import com.bitgirder.io.IoTestFactory;

import com.bitgirder.test.Test;

import java.nio.ByteBuffer;

import java.nio.channels.FileChannel;

@Test
final
class V1IoTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    @Test
    private
    void
    testBasicFileInfo()
        throws Exception
    {
        FileWrapper fw = IoTestFactory.createTempFile();
        int len = 10;

        FileChannel fc = fw.openWriteChannel();
        ByteBuffer bb = ByteBuffer.allocate( len );
        while ( bb.hasRemaining() ) fc.write( bb );
        fc.close();

        BasicFileInfo info = V1Io.basicFileInfoFor( fw );

        state.equalInt( len, info.size().getIntByteCount() );

        state.equal( 
            fw.getFile().lastModified(), info.mtime().getTimeInMillis() );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = "\\QFile.isFile() is false for \\E.+" )
    private
    void
    testBasicFileInfoForFailsWithDir()
        throws Exception
    {
        V1Io.basicFileInfoFor( DirWrapper.systemTmpdir().getFile() );
    }
}
