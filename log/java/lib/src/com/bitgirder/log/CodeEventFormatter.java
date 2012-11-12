package com.bitgirder.log;

public
interface CodeEventFormatter
{
    public
    void
    appendFormat( StringBuilder sb,
                  CodeEvent ev );
}
