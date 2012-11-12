package com.bitgirder.crypto;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class CryptoConstants
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static String KEY_STORE_TYPE_JCEKS = "jceks";
    final static String ALGO_DES_EDE = "DESede";
    final static String ALGO_DES = "DES";
}
