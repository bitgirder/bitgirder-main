package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class MingleBinaryConstants
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    final static byte TC_NIL = (byte) 0x00;
    final static byte TC_ID = (byte) 0x01;
    final static byte TC_NS = (byte) 0x02;
    final static byte TC_DECL_NM = (byte) 0x03;
    final static byte TC_QN = (byte) 0x04;
    final static byte TC_ATOM_TYP = (byte) 0x05;
    final static byte TC_LIST_TYP = (byte) 0x06;
    final static byte TC_NULLABLE_TYP = (byte) 0x07;
    final static byte TC_REGEX_RESTRICT = (byte) 0x08;
    final static byte TC_RANGE_RESTRICT = (byte) 0x09;
    final static byte TC_BOOL = (byte) 0x0a;
    final static byte TC_STRING = (byte) 0x0b;
    final static byte TC_INT32 = (byte) 0x0c;
    final static byte TC_INT64 = (byte) 0x0d;
    final static byte TC_UINT32 = (byte) 0x0e;
    final static byte TC_UINT64 = (byte) 0x0f;
    final static byte TC_FLOAT32 = (byte) 0x10;
    final static byte TC_FLOAT64 = (byte) 0x11;
    final static byte TC_TIME_RFC3339 = (byte) 0x12;
    final static byte TC_BUFFER = (byte) 0x13;
    final static byte TC_ENUM = (byte) 0x14;
    final static byte TC_SYM_MAP = (byte) 0x15;
    final static byte TC_MAP_PAIR = (byte) 0x16;
    final static byte TC_STRUCT = (byte) 0x17;
    final static byte TC_LIST = (byte) 0x19;
    final static byte TC_END = (byte) 0x1a;
    final static byte TC_ID_PATH = (byte) 0x1b;
    final static byte TC_ID_PATH_LIST_NODE = (byte) 0x1c;
}
