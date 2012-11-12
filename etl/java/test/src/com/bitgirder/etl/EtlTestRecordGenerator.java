package com.bitgirder.etl;

public
interface EtlTestRecordGenerator< V >
{
    public
    V
    next( long indx )
        throws Exception;
}
