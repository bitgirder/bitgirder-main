package com.bitgirder.mingle.bincodec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class MingleBinaryCodecConstants
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleBinaryCodecConstants() {}

    final static byte TYPE_CODE_END = (byte) 0;
    final static byte TYPE_CODE_BOOLEAN = (byte) 1;
    final static byte TYPE_CODE_DOUBLE = (byte) 3;
    final static byte TYPE_CODE_ENUM = (byte) 4;
    final static byte TYPE_CODE_FLOAT = (byte) 5;
    final static byte TYPE_CODE_INT32 = (byte) 6;
    final static byte TYPE_CODE_UINT32 = (byte) 7;
    final static byte TYPE_CODE_INT64 = (byte) 8;
    final static byte TYPE_CODE_UINT64 = (byte) 9;
    final static byte TYPE_CODE_RFC3339_STR = (byte) 11;
    final static byte TYPE_CODE_RFC3339_BIN = (byte) 12;
    final static byte TYPE_CODE_BUFFER = (byte) 13;
    final static byte TYPE_CODE_UTF8_STRING = (byte) 14;
    final static byte TYPE_CODE_LIST = (byte) 15;
    final static byte TYPE_CODE_STRUCT = (byte) 16;
    final static byte TYPE_CODE_EXCEPTION = (byte) 17;
    final static byte TYPE_CODE_SYMBOL_MAP = (byte) 18;
    final static byte TYPE_CODE_NULL = (byte) 19;
}
