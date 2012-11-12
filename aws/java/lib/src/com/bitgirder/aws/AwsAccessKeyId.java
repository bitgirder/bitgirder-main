package com.bitgirder.aws;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

public
final
class AwsAccessKeyId
extends TypedString< AwsAccessKeyId >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public AwsAccessKeyId( CharSequence cs ) { super( cs, "cs" ); }
}
