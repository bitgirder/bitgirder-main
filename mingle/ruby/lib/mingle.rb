require 'bitgirder/core'
include BitGirder::Core

require 'bitgirder/io'

require 'bigdecimal'
require 'forwardable'

module Mingle
    
# For processing \uXXXX escapes in string tokens
USE_ICONV = ! "".respond_to?( :encode )
require 'iconv' if USE_ICONV

class MingleValue < BitGirderClass
end

class MingleNull < MingleValue

    private_class_method :new

    INSTANCE = self.send( :new )
end

# Note: it's important that this class be defined before
# things like MingleNamespace, MingleIdentifier, etc, which register parse
# handlers for values of type MingleString.
class MingleString < MingleValue

    bg_attr :str

    extend Forwardable
    
    def_delegators :@str, :[], :=~, :to_sym, :split

    include Comparable

    def initialize( str )
        @str = not_nil( str, "str" ).to_s.dup
    end

    public
    def <=>( other )
        
        if other.class == self.class
            @str <=> other.str
        else
            raise TypeError, "Not a #{self.class}: #{other.class}"
        end
    end

    public
    def to_s
        @str.to_s
    end

    public
    def inspect
        to_s.inspect
    end

    public
    def ==( other )
        
        return true if other.equal?( self )
        return false unless other.is_a?( MingleString )

        other_str = other.instance_variable_get( :@str )
        @str == other_str
    end

    public
    def eql?( other )
        self == other
    end

    public
    def hash
        @str.hash
    end

    public
    def to_i
        @str.to_i
    end
end

class MingleBoolean < MingleValue
    
    private_class_method :new

    extend Forwardable

    def initialize( b )
        @b = b
    end

    public
    def to_s
        @b.to_s
    end

    public
    def inspect
        to_s.inspect
    end

    public
    def as_boolean
        @b
    end

    alias to_b as_boolean
    alias to_bool as_boolean
    alias to_boolean as_boolean

    public
    def ==( other )
        other.is_a?( MingleBoolean ) && @b == other.as_boolean
    end

    def_delegator :@b, :hash
    alias eql? ==
    
    TRUE = new( true )
    FALSE = new( false )

    def self.for_boolean( b )
        b ? TRUE : FALSE
    end
end

class MingleNumber < MingleValue

    include Comparable
    extend Forwardable

    attr_reader :num

    def initialize( num, convert_meth = nil )
        
        not_nil( num, :num )
        @num = convert_meth ? num.send( convert_meth ) : num
    end

    public
    def to_s
        @num.to_s
    end

    public
    def inspect
        to_s.inspect
    end

    public
    def ==( other )
        other.class == self.class && other.num == @num
    end

    def_delegator :@num, :hash
    alias eql? ==

    public
    def <=>( other )
        
        if other.is_a?( MingleNumber )
            @num <=> other.num
        else
            raise TypeError, other.class.to_s
        end
    end

    public
    def to_i
        @num.to_i
    end

    public
    def to_f
        @num.to_f
    end
end

class MingleIntegerImpl < MingleNumber

    def self.can_hold?( num )
        num >= self::MIN_NUM && num <= self::MAX_NUM
    end

    def initialize( num )
        super( num, :to_i )
    end
end

class MingleFloatingPointImpl < MingleNumber
    
    def initialize( num, fmt )
        super( [ num.to_f ].pack( fmt ).unpack( fmt )[ 0 ] )
    end
end

class MingleFloat64 < MingleFloatingPointImpl
    def initialize( num )
        super( num, 'd' )
    end
end

class MingleFloat32 < MingleFloatingPointImpl
    
    def initialize( num )
        super( num, 'f' )
    end
end

class MingleInt32 < MingleIntegerImpl

    MAX_NUM = ( 2 ** 31 ) - 1
    MIN_NUM = -( 2 ** 31 )
end

class MingleInt64 < MingleIntegerImpl
    
    MAX_NUM = ( 2 ** 63 ) - 1
    MIN_NUM = -( 2 ** 63 )
end

class MingleUint32 < MingleIntegerImpl
    
    MAX_NUM = ( 2 ** 32 ) - 1
    MIN_NUM = 0
end

class MingleUint64 < MingleIntegerImpl
    
    MAX_NUM = ( 2 ** 64 ) - 1
    MIN_NUM = 0
end

class MingleBuffer < MingleValue
    
    extend Forwardable

    attr_reader :buf

    class EncodingError < StandardError; end

    private
    def process_buffer_encoding( buf, encode_mode )
        
        enc_bin = Encoding::BINARY

        if ( enc = buf.encoding ) != enc_bin

            case encode_mode

                when :copy then buf = buf.encode( enc_bin )
                when :in_place then buf = buf.encode!( enc_bin )

                when :none 
                    raise EncodingError, 
                          "Encoding should be binary (got : #{enc})"

                else raise "Invalid encode mode: #{encode_mode}"
            end
        end

        buf
    end

    def initialize( buf, encode_mode = :none )

        not_nil( buf, "buf" )
        
        @buf = RubyVersions.when_19x( buf ) do
            process_buffer_encoding( buf, encode_mode )
        end
    end

    public
    def ==( other )
        other.is_a?( MingleBuffer ) && other.buf == @buf
    end

    def_delegator :@buf, :hash
    alias eql? ==
end

class MingleTimestamp < MingleValue
 
    extend BitGirder::Core::BitGirderMethods
    include Comparable

    extend Forwardable

    require 'time'

    attr_reader :time

    # Uses iso8601 serialize --> parse to make a copy of the supplied time
    # unless make_copy is false, in which case time is used directly by this
    # instance
    def initialize( time, make_copy = true )
        
        not_nil( time, "time" )
        @time = ( make_copy ? Time.iso8601( time.iso8601( 9 ) ) : time ).utc
    end

    def self.now
        new( Time.now, false )
    end

    class Rfc3339FormatError < StandardError; end

    def self.rfc3339( str )

        begin
            new( Time.iso8601( not_nil( str, :str ).to_s ), false )
        rescue ArgumentError => ae
            
            if ae.message =~ /^invalid date: /
                raise Rfc3339FormatError.new( ae.message )
            else
                raise ae
            end
        end
    end

    def self.from_seconds( secs )
        
        not_nil( secs, :secs )
        new( Time.at( secs ), false )
    end

    # Impl :note => simply calling Time.at( ms / 1000.0 ) doesn't work as we
    # might want, since it ends up passing a Float to Time.at() which apparently
    # performs more calculations or otherwise leads to a time which is close to
    # but not precisely the result of the division. To illustrate:
    #
    #  irb(main):013:0> Time.at( 1299534304123 / 1000.0 ).iso8601( 9 )
    #  => "2011-03-07T13:45:04.122999907-08:00"
    #
    # while the algorithm we use, which uses integral values only, gives the 123
    # fractional value as expected:
    #
    # irb(main):014:0> Time.at( 1299534304123 / 1000, ( 1299534304123 % 1000 ) *
    # 1000 ).iso8601( 9 )
    # => "2011-03-07T13:45:04.123000000-08:00"
    #
    def self.from_millis( ms )

        not_nil( ms, :ms )

        secs = ms / 1000
        usec = ( ms % 1000 ) * 1000

        new( Time.at( secs, usec ), false )
    end

    public
    def rfc3339
        @time.iso8601( 9 )
    end

    alias to_s rfc3339

    public
    def to_i
        @time.to_i
    end

    public
    def to_f
        @time.to_f
    end

    public
    def ==( other )
        other.is_a?( MingleTimestamp ) && other.rfc3339 == rfc3339
    end

    def_delegator :@time, :hash
    alias eql? ==

    public
    def <=>( other )
        
        if other.is_a?( MingleTimestamp ) 
            @time <=> other.time
        else
            raise TypeError, other.class.to_s
        end
    end
end 

class MingleList < MingleValue
    
    extend Forwardable
    include Enumerable

    def_delegators :@arr, :each, :[], :join, :empty?, :size

    def initialize( obj )
        @arr = obj.map { |elt| MingleModels.as_mingle_value( elt ) }.freeze
    end

    public
    def to_s
        @arr.to_s
    end

    public
    def to_a
        Array.new( @arr )
    end

    public
    def ==( other )
        other.is_a?( MingleList ) &&
            other.instance_variable_get( :@arr ) == @arr
    end

    public
    def +( coll )
        
        not_nil( coll, :coll )

        case coll

            when MingleList 
                MingleList.new( @arr + coll.instance_variable_get( :@arr ) )

            when Array then self + MingleList.new( coll )

            else 
                raise "Operation '+' not supported for objects of " \
                      "type #{coll.class}"
        end
    end
end

class ParseLocation < BitGirderClass
    
    bg_attr :source, :default => "<>"
    bg_attr :line, :validation => :nonnegative, :default => 1
    bg_attr :col, :validation => :nonnegative

    public
    def to_s
        "[#@source, line #@line, col #@col]"
    end
end

class MingleParseError < BitGirderError

    bg_attr :err
    bg_attr :loc

    public
    def to_s
        "#@loc: #@err"
    end 
end

class RestrictionTypeError < StandardError; end

PARSED_TYPES = [ MingleString, String, Symbol ]

ID_STYLE_LC_CAMEL_CAPPED = :lc_camel_capped
ID_STYLE_LC_UNDERSCORE = :lc_underscore
ID_STYLE_LC_HYPHENATED = :lc_hyphenated

ID_STYLES = [ 
    ID_STYLE_LC_CAMEL_CAPPED, 
    ID_STYLE_LC_UNDERSCORE,
    ID_STYLE_LC_HYPHENATED
]

class SpecialToken < BitGirderClass
    
    bg_attr :val

    COLON = new( :val => ":" )
    ASPERAND = new( :val => "@" )
    PERIOD = new( :val => "." )
    FORWARD_SLASH = new( :val => "/" )
    PLUS = new( :val => "+" )
    MINUS = new( :val => "-" )
    ASTERISK = new( :val => "*" )
    QUESTION_MARK = new( :val => "?" )
    TILDE = new( :val => "~" )
    OPEN_BRACKET = new( :val => "[" )
    CLOSE_BRACKET = new( :val => "]" )
    OPEN_PAREN = new( :val => "(" )
    CLOSE_PAREN = new( :val => ")" )
    COMMA = new( :val => "," )

    TOK_CHARS = ":@./+-*?~[](),"

    public
    def hash
        @val.hash
    end

    public
    def to_s
        @val.to_s
    end
end

class WhitespaceToken < BitGirderClass
    
    bg_attr :ws

    public
    def to_s
        @ws.to_s.inspect
    end
end

module Chars

    module_function

    SIMPLE_ESCAPE_VALS = "\n\r\t\f\b"
    SIMPLE_ESCAPE_STRS = '\n\r\t\f\b'

    def ctl_char?( ch )
        ( 0x00 ... 0x20 ).include?( ch.ord )
    end

    def get_simple_escape( ch )
        
        if i = SIMPLE_ESCAPE_VALS.index( ch.chr )
            SIMPLE_ESCAPE_STRS[ 2 * i, 2 ]
        else
            nil
        end
    end

    def external_form_of( val )

        res = RubyVersions.when_19x( '"' ) { |s| s.encode!( "binary" ) }

        val.each_byte do |b|
            
            case
            when Chars.ctl_char?( b )
                if s = Chars.get_simple_escape( b )
                    res << s
                else
                    res << sprintf( "\\u%04X", b )
                end
            when b == ?".ord || b == ?\\.ord then res << "\\" << b
            else res << b
            end
        end

        RubyVersions.when_19x( res << '"' ) { |s| s.force_encoding( "utf-8" ) }
    end
end

class StringToken < BitGirderClass

    bg_attr :val

    public
    def hash
        @val.hash
    end

    public
    def external_form
        Chars.external_form_of( @val )
    end

    alias to_s external_form
end

class NumericToken < BitGirderClass
 
    bg_attr :int
    bg_attr :frac, :default => ""
    bg_attr :exp, :default => ""
    bg_attr :exp_char, :default => ""

    public
    def hash
        [ @int, @frac, @exp, @exp_char ].hash
    end

    public
    def external_form
        
        res = @int.dup
        ( res << "." << @frac ) unless @frac.empty?
        ( res << @exp_char << @exp ) unless @exp_char.empty?

        res
    end

    alias to_s external_form

    public
    def integer?
        @exp_char.empty? && @frac.empty?
    end
end

class MingleLexer < BitGirderClass

    bg_attr :io

    map_instance_of( String ) do |s| 
        s = s.dup if s.frozen?
        self.new( :io => StringIO.new( s ) )
    end

    LC_ALPHA = ( ?a .. ?z )
    IDENT_SEPS = [ ?-, ?_ ]
    UC_ALPHA = ( ?A .. ?Z )
    DIGIT = ( ?0 .. ?9 )
    UC_HEX = ( ?A .. ?F )
    LC_HEX = ( ?a .. ?f )
    LEAD_SURROGATE = ( 0xD800 ... 0xDC00 )
    TRAIL_SURROGATE = ( 0xDC00 ... 0xE000 )

    private
    def impl_initialize
        @line, @col = 1, 0
    end

    public
    def eof?
        @io.eof?
    end

    public
    def create_loc( col_adj = 0 )
        ParseLocation.new( :col => @col + col_adj, :line => @line )
    end

    private
    def impl_fail_parse( msg, loc )
        raise MingleParseError.new( :err => msg, :loc => loc )
    end

    private
    def fail_parse( msg )
        impl_fail_parse( msg, create_loc )
    end

    private
    def fail_parsef( *argv )
        fail_parse( sprintf( *argv ) )
    end

    private
    def fail_unexpected_end( msg = "Unexpected end of input" )
        
        @col += 1 if eof?
        fail_parse( msg )
    end

    # For compatibility and ease of asserting error messages, we make sure this
    # converts \t --> "\t", \n --> "\n", etc, and otherwise converts 0x01 -->
    # "\x01" (even though ruby 1.9x would yield "\u0001")
    private
    def inspect_char( ch )
        case
        when ch == ?\n then '"\n"'
        when ch == ?\t then '"\t"'
        when ch == ?\f then '"\f"'
        when ch == ?\r then '"\r"'
        when ch == ?\b then '"\b"'
        when Chars.ctl_char?( ch ) then sprintf( '"\x%02X"', ch.ord )
        else ch.chr.inspect
        end
    end

    private
    def err_ch( ch, ch_desc = nil )

        if ch
            ch_desc ||= inspect_char( ch )
            sprintf( "#{ch_desc} (0x%02X)", ch.ord ) 
        else
            "END"
        end
    end

    private
    def get_char( fail_on_eof = false )
        
        if ch = @io.getc

            if ch == ?\n
                @unread_col, @col = @col, 0
                @line += 1
            else
                @col += 1
            end

            ch
        else
            fail_parse( "Unexpected end of input" ) if fail_on_eof
        end
    end

    # Okay to call with nil (okay to unget EOF)
    private
    def unget_char( ch )
        
        if ch

            @io.ungetc( ch )
    
            if ch == ?\n
                @line, @col = @line - 1, @unread_col
            else
                @col -= 1
            end
        end
    end

    private
    def peek_char
        get_char.tap { |ch| unget_char( ch ) }
    end

    private
    def poll_chars( *expct )

        if expct.include?( ch = get_char )
            ch
        else
            unget_char( ch )
            nil
        end
    end

    private
    def ident_start?( ch )
        LC_ALPHA.include?( ch )
    end

    private
    def ident_part_char?( ch )
        [ LC_ALPHA, DIGIT ].find { |rng| rng.include?( ch ) }
    end

    private
    def ident_part_sep?( ch )
        [ IDENT_SEPS, UC_ALPHA ].find { |rng| rng.include?( ch ) }
    end

    private
    def sep_char_for( styl )
        case styl
        when ID_STYLE_LC_HYPHENATED then ?-
        when ID_STYLE_LC_UNDERSCORE then ?_
        else nil
        end
    end

    private
    def can_trail?( styl )
        styl == ID_STYLE_LC_UNDERSCORE || styl == ID_STYLE_LC_HYPHENATED
    end

    private
    def read_ident_part_start( styl, expct )
        
        ch, res = get_char, nil

        if styl == ID_STYLE_LC_CAMEL_CAPPED
            res = ch.chr.downcase if UC_ALPHA.include?( ch )
        else
            res = ch if ident_start?( ch )
        end

        unless res
            if expct
                fail_parse "Illegal start of identifier part: #{err_ch( ch )}"
            else
                unget_char( ch ) 
            end
        end

        res
    end

    private
    def read_ident_sep( ch, styl )
 
        if styl
            if ch == sep_char_for( styl )
                if eof? && can_trail?( styl )
                    fail_unexpected_end( "Empty identifier part" ) 
                end
            else
                unget_char( ch ) 
            end
        else
            case ch
            when ?- then styl = ID_STYLE_LC_HYPHENATED
            when ?_ then styl = ID_STYLE_LC_UNDERSCORE
            else 
                styl = ID_STYLE_LC_CAMEL_CAPPED
                unget_char( ch )
            end
        end

        styl
    end

    private
    def read_ident_part_tail( part, styl )
 
        part_done = false

        begin

            ch = get_char
            case
            when ident_part_char?( ch ) then part << ch
            when ident_part_sep?( ch ) 
                styl, part_done = read_ident_sep( ch, styl ), true
            else 
                part_done, id_done = true, true
                unget_char( ch )
            end

        end until part_done

        [ styl, id_done ]
    end

    private
    def read_ident_part( styl, expct )
        
        part, id_done = "", false

        if ch = read_ident_part_start( styl, expct )

            part << ch
            styl, id_done = read_ident_part_tail( part, styl )
        end

        [ part, styl, part.empty? || id_done ]
    end

    private
    def read_ident( styl = nil )
        
        parts = []
        
        begin
            unless eof?
                expct = parts.empty? || can_trail?( styl )
                part, styl, id_done = read_ident_part( styl, expct )
                parts << part unless part.empty?
            end

        end until id_done || eof?

        fail_unexpected_end( "Empty identifier" ) if parts.empty?
        MingleIdentifier.send( :new, :parts => parts )
    end

    private
    def decl_nm_start?( ch )
        UC_ALPHA.include?( ch )
    end

    private
    def decl_nm_char?( ch )
        [ UC_ALPHA, LC_ALPHA, DIGIT ].find { |rng| rng.include?( ch ) }
    end

    private
    def read_decl_type_name
        
        fail_unexpected_end( "Empty type name" ) if eof?

        if decl_nm_start?( ch = get_char )
            res = ch.chr
        else
            fail_parse( "Illegal type name start: #{err_ch( ch )}" )
        end

        begin
            if decl_nm_char?( ch = get_char )
                res << ch
            else
                unget_char( ch )
                ch = nil
            end
        end while ch

        DeclaredTypeName.send( :new, :name => res )
    end

    private
    def special_char?( ch )
        ch && SpecialToken::TOK_CHARS.index( ch )
    end

    private
    def read_special
        SpecialToken.new( :val => get_char.chr )
    end

    private
    def whitespace?( ch )
        ch && " \n\r\t".index( ch )
    end

    private
    def read_whitespace
        
        ws = ""

        begin
            if whitespace?( ch = get_char )
                ws << ch
            else
                unget_char( ch )
                ch = nil
            end
        end while ch

        WhitespaceToken.new( :ws => ws )
    end

    private
    def hex_char?( ch )
        [ DIGIT, UC_HEX, LC_HEX ].find { |rng| rng.include?( ch ) }
    end

    private
    def new_bin_str
        RubyVersions.when_19x( "" ) { |s| s.encode!( "binary" ) }
    end

    private
    def read_utf16_bytes

        Array.new( 2 ) do
            
            s = ""

            2.times do
                if hex_char?( ch = get_char )
                    s << ch
                else
                    fail_parse( "Invalid hex char in escape: #{err_ch( ch )}" )
                end
            end

            s.to_i( 16 )
        end
    end

    private
    def surrogate?( hi, lo, rng )
        rng.include?( ( hi << 8 ) + lo )
    end

    private
    def escape_utf16( bin )
        
        res = ""
        
        unless bin.size % 2 == 0
            raise "Bin string size #{bin.size} not a multiple of 4 bytes"
        end

        ( bin.size / 2 ).times do |i|
            res << sprintf( "\\u%04X", bin[ 2 * i, 2 ].unpack( "n" )[ 0 ] )
        end

        res
    end

    private
    def read_trail_surrogate( bin )

        tmpl = "Expected trailing surrogate, found: %s"

        unless ( ch = get_char( true ) ) == ?\\
            impl_fail_parse( sprintf( tmpl, err_ch( ch ) ), create_loc )
        end

        unless ( ch = get_char( true ) ) == ?u
            impl_fail_parse( sprintf( tmpl, "\\#{ch.chr}" ), create_loc( -1 ) )
        end

        hi, lo = read_utf16_bytes
        bin << hi << lo

        unless surrogate?( hi, lo, TRAIL_SURROGATE )
            msg = "Invalid surrogate pair #{escape_utf16( bin )}"
            impl_fail_parse( msg, create_loc( -11 ) )
        end
    end

    private
    def read_utf16_escape( dest )
        
        bin = new_bin_str

        hi, lo = read_utf16_bytes
        bin << hi << lo

        if surrogate?( hi, lo, LEAD_SURROGATE )
            read_trail_surrogate( bin ) 
        elsif surrogate?( hi, lo, TRAIL_SURROGATE )
            msg = "Trailing surrogate with no lead: #{escape_utf16( bin )}"
            impl_fail_parse( msg, create_loc( -5 ) )
        end

        if USE_ICONV
            dest << Iconv.conv( "utf-8", "utf-16be", bin )
        else
            dest << bin.encode!( "utf-8", "utf-16be" )
        end
    end

    private
    def read_escaped_char( dest )
        
        case ch = get_char
        when ?n then dest << "\n"
        when ?t then dest << "\t"
        when ?f then dest << "\f"
        when ?r then dest << "\r"
        when ?b then dest << "\b"
        when ?\\ then dest << "\\"
        when ?" then dest << "\""
        when ?u then read_utf16_escape( dest )
        else fail_parse( "Unrecognized escape: #{err_ch( ch, "\\#{ch.chr}" )}" )
        end
    end

    private
    def append_string_tok( dest, ch )
 
        if Chars.ctl_char?( ch )

            unget_char( ch ) # To reset line num in case we read \n
            msg = "Invalid control character in string literal: #{err_ch( ch )}"
            impl_fail_parse( msg, create_loc( 1 ) )
        else
            dest << ch
        end
    end

    private
    def read_string
        
        unless ( ch = get_char ) == ?"
            fail_parse( "Expected string start, saw #{err_ch( ch )}" )
        end

        res = RubyVersions.when_19x( "" ) { |s| s.encode!( "utf-8" ) }

        begin
            case ch = get_char
            when nil then fail_parse( "Unterminated string literal" )
            when ?\\ then read_escaped_char( res )
            when ?" then nil
            else append_string_tok( res, ch )
            end
        end until ch == ?"

        StringToken.new( :val => res )
    end

    private
    def starts_num?( ch )
        DIGIT.include?( ch )
    end

    private
    def read_dig_str( err_desc, *ends )
        
        res = ""

        begin
            if DIGIT.include?( ch = get_char )
                res << ch
            else
                if [ nil, ?e, ?E ].include?( ch ) || special_char?( ch )
                    unget_char( ch )
                    ch = nil
                else
                    fail_parse( 
                        "Unexpected char in #{err_desc}: #{err_ch( ch )}" )
                end
            end
        end while ch

        fail_parse( "Number has empty or invalid #{err_desc}" ) if res.empty? 

        res
    end

    private
    def read_num_exp( opts )
        
        if [ ?e, ?E ].include?( ch = get_char )

            opts[ :exp_char ] = ch.chr

            opts[ :exp ] = 
                ( poll_chars( ?-, ?+ ) == ?- ? "-" : "" ) + 
                read_dig_str( "exponent" )
        else
            if ch == nil || whitespace?( ch ) || 
                   ( ch != ?. && special_char?( ch ) )
                unget_char( ch )
            else
                fail_parse( 
                    "Expected exponent start or num end, found: " +
                    err_ch( ch )
                )
            end
        end
    end

    private
    def read_number
        
        opts = {}

        opts[ :int ] = read_dig_str( "integer part" )
        opts[ :frac ] = read_dig_str( "fractional part" ) if poll_chars( ?. )
        read_num_exp( opts )

        NumericToken.new( opts )
    end

    # Note about the case statement: the typ based checks need to fire before
    # char ones so that if, for example, typ is DeclaredTypeName and the input
    # is 'a', we will fail as a bad type name rather than returning the
    # identifier 'a'
    public
    def read_token( typ = nil )

        # Don't peek -- do get/unget so we get a true loc
        ch = get_char
        loc = create_loc
        unget_char( ch )

        case 
        when typ == StringToken then res = read_string
        when typ == NumericToken then res = read_number
        when typ == MingleIdentifier then res = read_ident
        when typ == DeclaredTypeName then res = read_decl_type_name
        when ident_start?( ch ) then res = read_ident
        when decl_nm_start?( ch ) then res = read_decl_type_name
        when special_char?( ch ) then res = read_special
        when whitespace?( ch ) then res = read_whitespace
        when ch == ?" then res = read_string
        when starts_num?( ch ) then res = read_number
        else fail_parsef( "Unrecognized token: #{err_ch( get_char )}" )
        end

        [ res, loc ]
    end

    public
    def expect_token( typ = nil )
        case
        when typ == nil then read_token || fail_unexpected_end
        when typ == StringToken || typ == NumericToken ||
             typ == MingleIdentifier || typ == DeclaredTypeName 
            read_token( typ )
        else raise "Unhandled token expect type: #{typ}"
        end
    end
end

class MingleParser < BitGirderClass
 
    bg_attr :lexer

    QUANTS = [ 
        SpecialToken::QUESTION_MARK, 
        SpecialToken::PLUS,
        SpecialToken::ASTERISK
    ]
    
    RANGE_OPEN = [ SpecialToken::OPEN_PAREN, SpecialToken::OPEN_BRACKET ]
    RANGE_CLOSE = [ SpecialToken::CLOSE_PAREN, SpecialToken::CLOSE_BRACKET ]

    private
    def loc
        @saved ? @saved[ 1 ] : @lexer.create_loc
    end

    private
    def next_loc
        @saved ? @saved[ 1 ] : @lexer.create_loc( 1 )
    end

    private
    def eof?
        @saved == nil && @lexer.eof?
    end

    private
    def read_tok( typ = nil )

        pair, @saved = @saved, nil
        pair || ( @lexer.eof? ? [ nil, nil ] : @lexer.read_token( typ ) )
    end

    private
    def unread_tok( tok, loc )
        
        pair = [ tok, loc ]
        raise "Attempt to unread #{pair} while @saved is #@saved" if @saved
        @saved = tok ? pair : nil
    end

    private
    def peek_tok( typ = nil )
        read_tok.tap { |pair| unread_tok( *pair ) }
    end

    private
    def fail_parse( msg, err_loc = loc )
        raise MingleParseError.new( :err => msg, :loc => err_loc )
    end

    private
    def fail_unexpected_token( desc )
        
        tok, err_loc = read_tok

        unless tok
            tok, err_loc = "END", @lexer.create_loc( 1 )
        end

        fail_parse( "Expected #{desc} but found: #{tok}", err_loc )
    end

    private
    def skip_ws
        begin
            tok, loc = read_tok
            unless tok.is_a?( WhitespaceToken )
                unread_tok( tok, loc )
                tok = nil
            end
        end while tok
    end

    public
    def check_trailing
        
        res = yield if block_given?
        
        unless eof?

            tok, _ = read_tok
            fail_parse( "Unexpected token: #{tok}" )
        end

        res
    end

    private
    def poll_special_loc( *specs )
        
        tok, loc = read_tok 

        unless specs.include?( tok )
            unread_tok( tok, loc )
            tok, loc = nil, nil
        end

        [ tok, loc ]
    end

    private
    def expect_special_loc( desc, *spec )
        
        res = poll_special_loc( *spec ) 

        if tok = res[ 0 ] 
            res
        else
            fail_unexpected_token( desc )
        end
    end

    %w{ poll expect }.each do |nm|
        define_method( :"#{nm}_special" ) do |*argv|
            self.send( :"#{nm}_special_loc", *argv )[ 0 ]
        end
    end

    # Does not allow for unread of its result
    private
    def expect_typed_loc( typ )
        
        if @saved
            if ( id = @saved[ 0 ] ).is_a?( typ )
                @saved.tap { @saved = nil }
            else
                fail_parse( "Expected identifier" )
            end
        else
            @lexer.expect_token( typ )
        end
    end

    private
    def expect_typed( typ )
        expect_typed_loc( typ )[ 0 ]
    end

    public
    def expect_number
 
        tok, _ = peek_tok

        if neg = ( tok == SpecialToken::MINUS )
            read_tok
        end

        ParsedNumber.new(
            :negative => neg,
            :num => expect_typed( NumericToken )
        )
    end

    public
    def expect_identifier
        expect_typed( MingleIdentifier )
    end

    public
    def expect_namespace
 
        parts = [ expect_identifier ]

        begin
            if colon = poll_special( SpecialToken::COLON )
                parts << expect_identifier
            end
        end while colon

        poll_special( SpecialToken::ASPERAND ) or
            fail_unexpected_token( "':' or '@'" )

        ver = expect_identifier
        
        MingleNamespace.send( :new, :parts => parts, :version => ver )
    end

    public
    def expect_declared_type_name
        expect_typed( DeclaredTypeName )
    end

    public
    def expect_qname
        
        ns = expect_namespace
        expect_special( "type path", SpecialToken::FORWARD_SLASH )
        nm = expect_declared_type_name

        QualifiedTypeName.send( :new, :namespace => ns, :name => nm )
    end

    public
    def expect_identified_name
        
        ns = expect_namespace

        names = []

        begin
            if tok = poll_special( SpecialToken::FORWARD_SLASH )
                names << expect_identifier
            end
        end while tok

        fail_parse( "Missing name", next_loc ) if names.empty?

        MingleIdentifiedName.send( :new, :namespace => ns, :names => names )
    end

    private
    def expect_type_name

        tok, _ = peek_tok
        
        case tok
        when MingleIdentifier then expect_qname
        when DeclaredTypeName then expect_declared_type_name
        else fail_unexpected_token( "identifier or declared type name" )
        end
    end

    private
    def resolve_type_name( nm )
        
        if res = QNAME_RESOLV_MAP[ nm ]
            res
        else
            nm
        end
    end

    private
    def fail_restriction_type( found, desc )
        raise RestrictionTypeError.new( 
            "Invalid target type for #{desc} restriction: #{found}" )
    end

    private
    def expect_pattern_restriction( nm )
        
        str, loc = expect_typed_loc( StringToken )

        if nm == QNAME_STRING
            begin
                RegexRestriction.new( :ext_pattern => str.val )
            rescue RegexpError => e
                raise RestrictionTypeError, "Invalid regex: #{e.message}"
            end
        else
            fail_restriction_type( nm, "regex" )
        end
    end

    # Expects one of the tokens in toks, and returns the pair [ true, loc ] or [
    # false, loc ] based on whether the expected pair had the token matching
    # close_test
    private
    def read_range_bound( err_desc, toks, close_test )

        pair = expect_special_loc( err_desc, *toks )

        [ pair[ 0 ] == close_test, pair[ 1 ] ]
    end

    private
    def check_allow_cast( val, typ, bound )

        err_desc = nil

        case val
        when StringToken
            err_desc = "string" if NUM_TYPES.include?( typ )
        when ParsedNumber
            if NUM_TYPES.include?( typ )
                if INT_TYPES.include?( typ ) && ( ! val.integer? )
                    err_desc = "decimal" 
                end
            else 
                err_desc = "number"
            end
        end

        if err_desc
            raise RestrictionTypeError.new( 
                "Got #{err_desc} as #{bound} value for range" )
        end
    end

    private
    def exec_cast( ms, typ )
        
        begin
            Mingle.cast_value( MingleString.new( ms ), typ )
        rescue MingleTimestamp::Rfc3339FormatError => e
            e = RestrictionTypeError.new( e.message ) if typ == TYPE_TIMESTAMP
            raise e
        end
    end

    private
    def cast_range_value( val, nm, bound )
        
        return nil if val.nil?

        typ = AtomicTypeReference.send( :new, :name => nm )

        unless COMPARABLE_TYPES.include?( typ )
            fail_restriction_type( typ, "range" )
        end

        check_allow_cast( val, typ, bound )

        # s is a StringToken or a ParsedNumber
        s = val.is_a?( StringToken ) ? val.val : val.external_form

        exec_cast( MingleString.new( s ), typ )
    end

    private
    def read_range_value( nm )
        
        tok, _ = peek_tok

        val = case
        when tok == SpecialToken::COMMA || RANGE_CLOSE.include?( tok ) then nil
        when tok == SpecialToken::MINUS || tok.is_a?( NumericToken )
            expect_number
        when tok.is_a?( StringToken ) then expect_typed( StringToken )
        else fail_unexpected_token( "range value" )
        end
    end

    private
    def check_range_restriction_syntax( opts )

        min_closed, min_sx, max_sx, max_closed = 
            opts.values_at( :min_closed, :min_syntax, :max_syntax, :max_closed )

        err_desc, loc =
            case
            when min_sx.nil? && max_sx.nil? && ( min_closed || max_closed )
                [ " ", opts[ :open_loc ] ]
            when min_sx.nil? && min_closed then [ " low ", opts[ :open_loc ] ]
            when max_sx.nil? && max_closed then [ " high ", opts[ :close_loc ] ]
            end

        fail_parse( "Infinite#{err_desc}range must be open", loc ) if loc
    end

    private
    def read_range_restriction_syntax( nm )
        
        res = {}

        res[ :min_closed ], res[ :open_loc ] = 
            read_range_bound( 
                "range open", RANGE_OPEN, SpecialToken::OPEN_BRACKET )
 
        res[ :min_syntax ] = read_range_value( nm )
        skip_ws
        expect_special( ",", SpecialToken::COMMA )
        res[ :max_syntax ] = read_range_value( nm )
        skip_ws

        res[ :max_closed ], res[ :close_loc ] = 
            read_range_bound( 
                "range close", RANGE_CLOSE, SpecialToken::CLOSE_BRACKET )
        
        res
    end

    private
    def adjacent_ints?( min, max )
        
        case min
        when MingleInt32, MingleInt64, MingleUint32, MingleUint64
            min.num.succ == max.num
        else false
        end
    end

    private
    def check_satisfiable_range( opts )
        
        min_closed, min, max, max_closed = 
            opts.values_at( :min_closed, :min, :max, :max_closed )

        # If min,max are nil this is an infinite range and therefore
        # satisifiable (bounds will have been checked to be open elsewhere)
        return if min.nil? && max.nil? 

        cmp, failed = min <=> max, false

        case
        when cmp == 0 then failed = ! ( min_closed && max_closed )
        when cmp > 0 then failed = true
        else
            unless min_closed || max_closed
                failed = adjacent_ints?( min, max )
            end
        end

        raise RestrictionTypeError, "Unsatisfiable range" if failed
    end

    private
    def expect_range_restriction( nm )
        
        opts = read_range_restriction_syntax( nm )
        check_range_restriction_syntax( opts )
 
        opts[ :min ] = cast_range_value( opts[ :min_syntax ], nm, "min" )
        opts[ :max ] = cast_range_value( opts[ :max_syntax ], nm, "max" )

        check_satisfiable_range( opts )

        RangeRestriction.new( opts )
    end

    private
    def expect_type_restriction( nm )

        tok, _ = peek_tok

        case 
        when tok.is_a?( StringToken ) then expect_pattern_restriction( nm )
        when RANGE_OPEN.include?( tok ) then expect_range_restriction( nm )
        else fail_unexpected_token( "type restriction" )
        end
    end

    private
    def expect_atomic_type_reference
        
        nm = resolve_type_name( expect_type_name )
        
        skip_ws
        restr = nil

        if poll_special( SpecialToken::TILDE )
            skip_ws
            restr = expect_type_restriction( nm )
        end

        AtomicTypeReference.send( :new, :name => nm, :restriction => restr )
    end

    private
    def poll_quants
        
        res = []

        begin
            if tok = poll_special( *QUANTS )
                res << tok
            end
        end while tok

        res
    end

    public
    def expect_type_reference
        
        at = expect_atomic_type_reference
        
        poll_quants.inject( at ) do |typ, quant|
            case quant
            when SpecialToken::ASTERISK
                ListTypeReference.send( 
                    :new, :element_type => typ, :allows_empty => true )
            when SpecialToken::PLUS
                ListTypeReference.send(
                    :new, :element_type => typ, :allows_empty => false )
            when SpecialToken::QUESTION_MARK
                NullableTypeReference.send( :new, :type => typ )
            end
        end
    end

    def self.for_string( s )
        self.new( :lexer => MingleLexer.as_instance( s ) )
    end

    def self.consume_string( s )
 
        p = self.for_string( s )
        yield( p ).tap { p.check_trailing }
    end
end

# Including classes must define impl_parse() as a class method; this module will
# add class methods parse()/get() which can take any of PARSED_TYPES and return
# an instance according to impl_parse. Also installs BitGirderClass instance
# handlers for PARSED_TYPES
module StringParser

    def self.included( cls )

        cls.class_eval do

            map_instance_of( *PARSED_TYPES ) { |s| self.impl_parse( s.to_s ) }

            def self.get( val )
                self.as_instance( val )
            end

            def self.parse( val )
                self.as_instance( val )
            end
        end
    end
end 

class ParsedNumber < BitGirderClass
    
    include StringParser
    extend Forwardable

    bg_attr :negative, :default => false
    bg_attr :num

    def_delegators :@num, :integer?

    public
    def hash
        [ @negative, @num ].hash
    end

    public
    def external_form
        ( @negative ? "-" : "" ) << @num.external_form
    end

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_number }
    end
end

class MingleIdentifier < BitGirderClass
    
    include StringParser

    bg_attr :parts
 
    private_class_method :new

    private
    def impl_initialize
        
        @parts.each_with_index do |part, idx|
            raise "Empty id part at index #{idx}" if part.size == 0
        end
    end

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_identifier }
    end
    
    public
    def format( fmt )

        not_nil( fmt, :fmt )

        case fmt

            when ID_STYLE_LC_HYPHENATED then @parts.join( "-" )

            when ID_STYLE_LC_UNDERSCORE then @parts.join( "_" )

            when ID_STYLE_LC_CAMEL_CAPPED
                @parts[ 0 ] + 
                @parts[ 1 .. -1 ].map do 
                    |t| t[ 0, 1 ].upcase + t[ 1 .. -1 ] 
                end.join

            else raise "Invalid format: #{fmt}"
        end
    end

    public
    def external_form
        format( :lc_hyphenated )
    end
    
    alias to_s external_form

    public
    def to_sym
        format( :lc_underscore ).to_sym
    end

    public
    def hash
        @parts.hash
    end

    def self.as_format_name( id )

        sym = self.get( id ).to_sym

        if ID_STYLES.include?( sym )
            sym
        else
            raise "Unknown or invalid identifier format: #{id} (#{id.class})"
        end
    end
end

class MingleNamespace < BitGirderClass

    include StringParser

    bg_attr :parts
    bg_attr :version

    private_class_method :new

    def self.create( opts )
        self.send( :new, opts )
    end

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_namespace }
    end

    public
    def format( id_styl )
        parts = @parts.map { |p| p.format( id_styl ) }
        ver = @version.format( id_styl )
        "#{parts.join( ":" )}@#{ver}"
    end

    public
    def external_form
        format( ID_STYLE_LC_CAMEL_CAPPED )
    end
    
    alias to_s external_form

    public
    def inspect
        to_s.inspect
    end

    public
    def ==( other )
        other.is_a?( MingleNamespace ) &&
            other.parts == @parts && other.version == @version
    end

    public
    def hash
        @parts.hash | @version.hash
    end

    public
    def eql?( other )
        self == other
    end
end

class DeclaredTypeName < BitGirderClass
    
    include StringParser

    bg_attr :name

    private_class_method :new

    def external_form
        @name.to_s
    end

    alias to_s external_form

    def hash
        @name.hash
    end

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_declared_type_name }
    end
end

class QualifiedTypeName < BitGirderClass
    
    include StringParser

    bg_attr :namespace, :processor => MingleNamespace
    bg_attr :name, :processor => DeclaredTypeName

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_qname }
    end

    public
    def hash
        [ @namespace, @name ].hash
    end

    public
    def external_form
        "#{@namespace.external_form}/#{@name.external_form}"
    end

    alias to_s external_form
end

class MingleTypeReference < BitGirderClass
 
    include StringParser
 
    private_class_method :new

    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_type_reference }
    end

    bg_abstract :external_form

    public
    def to_s
        external_form
    end
end

class RegexRestriction < BitGirderClass

    bg_attr :ext_pattern # Raw text of original input pattern
    
    attr_reader :regexp

    install_hash

    private
    def impl_initialize
        @regexp = Regexp.new( @ext_pattern )
    end

    public
    def external_form
        StringToken.new( :val => @ext_pattern ).external_form
    end
end

class RangeRestriction < BitGirderClass
 
    bg_attr :min_closed, :processor => :boolean
    bg_attr :min, :required => false
    bg_attr :max, :required => false
    bg_attr :max_closed, :processor => :boolean

    install_hash

    public
    def external_form
        
        res = @min_closed ? "[" : "("
        ( res << Mingle.quote_value( @min ) ) if @min
        res << ","
        ( res << Mingle.quote_value( @max ) ) if @max
        res << ( @max_closed ? "]" : ")" )

        res
    end
end

class AtomicTypeReference < MingleTypeReference
    
    bg_attr :name # A DeclaredTypeName or QualifiedTypeName
 
    bg_attr :restriction, :required => false

    def self.create( *argv )
        self.send( :new, *argv )
    end

    public
    def hash
        [ @name, @restriction ].hash
    end

    public
    def external_form
        
        res = @name.external_form.dup
        ( res << "~" << @restriction.external_form ) if @restriction

        res
    end
end

QNAME_RESOLV_MAP = {}

%w{ 
    Value 
    Boolean 
    Buffer 
    String 
    Int32 
    Int64 
    Uint32
    Uint64
    Float32 
    Float64 
    Timestamp 
    Enum
    SymbolMap 
    Struct 
    Null

}.each do |nm|
    
    uc = nm.upcase

    nm = DeclaredTypeName.send( :new, :name => nm )
    
    qn = const_set( :"QNAME_#{uc}", 
        QualifiedTypeName.send( :new,
            :namespace => MingleNamespace.send( :new,
                :parts => [
                    MingleIdentifier.send( :new, :parts => %w{ mingle } ),
                    MingleIdentifier.send( :new, :parts => %w{ core } )
                ],
                :version => MingleIdentifier.send( :new, :parts => %w{ v1 } )
            ),
            :name => nm
        )
    )

    QNAME_RESOLV_MAP[ nm ] = qn

    const_set( :"TYPE_#{uc}", AtomicTypeReference.send( :new, :name => qn ) )
end

NUM_TYPES = 
    %w{ INT32 INT64 UINT32 UINT64 FLOAT32 FLOAT64 }.map do |s| 
        const_get( :"TYPE_#{s}" )
    end

INT_TYPES = [ TYPE_INT32, TYPE_INT64, TYPE_UINT32, TYPE_UINT64 ]

COMPARABLE_TYPES = NUM_TYPES + [ TYPE_STRING, TYPE_TIMESTAMP ]

class ListTypeReference < MingleTypeReference
    
    bg_attr :element_type
    bg_attr :allows_empty, :processor => :boolean

    public
    def hash
        [ @element_type, @allows_empty ].hash
    end

    public
    def external_form
        "#{@element_type.external_form}#{@allows_empty ? "*" : "+"}"
    end
end

class NullableTypeReference < MingleTypeReference

    bg_attr :type

    public
    def hash
        @type.hash
    end

    public
    def external_form
        "#{@type.external_form}?"
    end
end

class MingleIdentifiedName < BitGirderClass
    
    include StringParser
    
    bg_attr :namespace, :processor => MingleNamespace

    bg_attr :names, 
            :list_validation => :not_empty,
            :processor => lambda { |arr|
                arr.map { |nm| MingleIdentifier.as_instance( nm ) }
            }
    
    def self.impl_parse( s )
        MingleParser.consume_string( s ) { |p| p.expect_identified_name }
    end

    public
    def hash
        [ @namespace, @names ].hash
    end

    public
    def external_form
        
        @names.inject( @namespace.format( ID_STYLE_LC_HYPHENATED ) ) do |s, nm|
            s << "/" << nm.format( ID_STYLE_LC_HYPHENATED )
        end
    end
end

class MingleTypedValue < MingleValue

    bg_attr :type, :processor => MingleTypeReference
end

class MingleEnum < MingleTypedValue
    
    bg_attr :value, :processor => MingleIdentifier
    
    public
    def to_s
        "#{@type.external_form}.#{@value.external_form}"
    end

    public
    def ==( other )
        other.is_a?( MingleEnum ) && 
            other.type == @type &&
            other.value == @value
    end
end

class MingleSymbolMap < MingleValue

    include Enumerable 

    extend Forwardable
    def_delegators :@map, 
        :size, :empty?, :each, :each_pair, :to_s, :to_hash, :keys
    
    class NoSuchKeyError < StandardError; end

    extend BitGirder::Core::BitGirderMethods

    private_class_method :new

    def self.create( map = {} )
        
        res = {}

        not_nil( map, "map" ).each_pair do |k, v| 
            
            mv = MingleModels.as_mingle_value( v )
            res[ MingleIdentifier::get( k ) ] = mv unless mv.is_a?( MingleNull )
        end

        new( res )
    end

    map_instance_of( Hash ) { |h| self.create( h ) }

    def initialize( map )
        @map = map.freeze
    end

    def get_map
        {}.merge( @map )
    end

    public
    def fields
        self
    end

    public
    def []( key )

        case not_nil( key, :key )

            when MingleIdentifier then @map[ key ]

            when String then self[ key.to_sym ]

            when Symbol
                if res = ( @vals_by_sym ||= {} )[ key ]
                    res
                else
                    res = self[ MingleIdentifier.get( key ) ]
                    @vals_by_sym[ key ] = res if res

                    res
                end
            
            else raise TypeError, "Unexpected key type: #{key.class}"
        end
    end
    
    alias get []

    public
    def expect( key )

        if ( res = self[ key ] ) == nil
            raise NoSuchKeyError, "Map has no value for key: #{key}"
        else
            res
        end
    end

    public
    def values_at( *arg )
        not_nil( arg, :arg ).map { |k| get( k ) }
    end

    public
    def ==( other )
 
        other.is_a?( MingleSymbolMap ) &&
            other.instance_variable_get( :@map ) == @map
    end

    # Util function for method_missing
    private
    def expect_one_arg( args )
            
        case sz = args.size
            when 0 then nil
            when 1 then args[ 0 ]
            else raise ArgumentError, "Wrong number of arguments (#{sz} for 1)"
        end
    end

    private
    def method_missing( meth, *args )
        
        case meth.to_s
    
            when /^(expect|get)_(mingle_[a-z][a-z\d]*(?:_[a-z][a-z\d]*)*)$/
                
                case val = send( $1.to_sym, expect_one_arg( args ) )
                    when nil then nil
                    else MingleModels.as_mingle_instance( val, $2.to_sym )
                end

            when /^(expect|get)_string$/
                s = send( :"#{$1}_mingle_string", *args ) and s.to_s

            when /^(expect|get)_int$/
                s = send( :"#{$1}_mingle_int64", *args ) and s.to_i

            when /^(expect|get)_timestamp$/
                s = send( :"#{$1}_mingle_timestamp", *args )

            when /^(expect|get)_boolean$/
                s = send( :"#{$1}_mingle_boolean", *args ) and s.to_bool

            else super
        end
    end

    EMPTY = new( {} )
end

class MingleStruct < MingleValue

    bg_attr :type, :processor => MingleTypeReference
    
    bg_attr :fields,
            :default => MingleSymbolMap::EMPTY,
            :processor => MingleSymbolMap
    
    public
    def to_s
        "#@type:#{fields}"
    end

    public
    def []( fld )
        @fields[ fld ]
    end
end

class GenericRaisedMingleError < StandardError

    include BitGirder::Core::BitGirderMethods
    extend Forwardable

    def_delegators :@me, :type, :fields, :[]

    def initialize( me, trace = nil )
        
        @me = not_nil( me, :me )

        super( "#{@me.type}: #{@me[ :message ]}" )
        set_backtrace( trace ) if trace
    end
end

module MingleModels

    require 'base64'

    extend BitGirder::Core::BitGirderMethods

    NUM_TYPES = {
        :mingle_int64 => MingleInt64,
        :mingle_int32 => MingleInt32,
        :mingle_uint32 => MingleUint32,
        :mingle_uint64 => MingleUint64,
        :mingle_float64 => MingleFloat64,
        :mingle_float32 => MingleFloat32,
    }

    module_function

    def create_coerce_error( val, targ_type )
        TypeError.new "Can't coerce value of type #{val.class} to #{targ_type}"
    end

    def as_mingle_string( val )
        
        case val

            when MingleString then val

            when MingleNumber, MingleBoolean, MingleTimestamp 
                MingleString.new( val.to_s )

            else raise create_coerce_error( val, MingleString )
        end
    end

    def impl_string_to_num( str, typ )
        
        if [ MingleInt64, MingleInt32, MingleUint32, MingleUint64 ].
           include?( typ )
            str.to_i
        elsif [ MingleFloat64, MingleFloat32 ].include?( typ )
            str.to_f
        else
            raise "Unhandled number target type: #{typ}"
        end
    end

    def as_mingle_number( val, num_typ )
        
        case val

            when num_typ then val

            when MingleString 
                num_typ.new( impl_string_to_num( val.to_s, num_typ ) )

            when MingleNumber then num_typ.new( val.num )

            else raise create_coerce_error( val, num_typ )
        end
    end

    def as_mingle_boolean( val )
        
        case val

            when MingleBoolean then val

            when MingleString
                case val.to_s.strip.downcase
                    when "true" then MingleBoolean::TRUE
                    when "false" then MingleBoolean::FALSE
                end

            else raise create_coerce_error( val, MingleBoolean )
        end
    end

    def as_mingle_buffer( val )

        not_nil( val, "val" )

        case val

            when MingleBuffer then val

            when MingleString 
                MingleBuffer.new( BitGirder::Io.strict_decode64( val.to_s ) )

            else raise create_coerce_error( val, MingleBuffer )
        end
    end

    def as_mingle_timestamp( val )
 
        not_nil( val, "val" )

        case val    

            when MingleTimestamp then val
            when MingleString then MingleTimestamp.rfc3339( val )
            when Time then MingleTimestamp.new( val )

            else raise create_coerce_error( val, MingleTimestamp )
        end
    end

    def impl_as_typed_value( val, typ )
        if val.is_a?( typ )
            val
        else
            sym = get_coerce_type_symbol( typ )
            raise create_coerce_error( val, typ )
        end
    end

    def as_mingle_struct( val )
        impl_as_typed_value( val, MingleStruct )
    end

    def as_mingle_list( val )
        impl_as_typed_value( val, MingleList )
    end

    def as_mingle_symbol_map( val )
        impl_as_typed_value( val, MingleSymbolMap )
    end

    def as_mingle_integer( num )
        
        case 
            when MingleInt32.can_hold?( num ) then MingleInt32.new( num )
            when MingleInt64.can_hold?( num ) then MingleInt64.new( num )
            when MingleUint64.can_hold?( num ) then MingleUint64.new( num )
            else raise "Number is out of range for mingle integer types: #{num}"
        end
    end

    def as_mingle_value( val )
 
        case val
            when MingleValue then val
            when String then MingleString.new( val )
            when Symbol then MingleString.new( val.to_s )
            when TrueClass then MingleBoolean::TRUE
            when FalseClass then MingleBoolean::FALSE
            when Integer then as_mingle_integer( val )
            when Float, BigDecimal, Rational then MingleFloat64.new( val )
            when Array then MingleList.new( val )
            when Hash then MingleSymbolMap.create( val )
            when Time then as_mingle_timestamp( val )
            when nil then MingleNull::INSTANCE
            else raise "Can't create mingle value for instance of #{val.class}"
        end
    end

    # No assertions or other type checking at this :level => if it's a sym we
    # assume it is of the form mingle_foo_bar where foo_bar is a coercable type;
    # if a class we assume it is of the form MingleFooBar that we can convert to
    # a symbol of the former form.
    def get_coerce_type_symbol( typ )
        
        case typ

            when Symbol then typ

            when Class

                base_name = typ.to_s.split( /::/ )[ -1 ] # Get last name part
                str =
                    base_name.gsub( /(^[A-Z])|([A-Z])/ ) do |m| 
                        ( $2 ? "_" : "" ) + m.downcase 
                    end

                str.to_sym
        
            else raise "Unexpected coerce type indicator: #{typ}"
        end
    end

    def as_mingle_instance( val, typ )
 
        type_sym = get_coerce_type_symbol( not_nil( typ, "typ" ) )

        if num_typ = NUM_TYPES[ type_sym ]
            as_mingle_number( val, num_typ )
        else
            send( :"as_#{type_sym}", val )
        end
    end

    def as_ruby_error( me, trace = nil )
        GenericRaisedMingleError.new( not_nil( me, :me ), trace )
    end

    def raise_as_ruby_error( me )
        raise as_ruby_error( me )
    end
end

class TypeCastError < BitGirderError
    
    bg_attr :actual
    bg_attr :expected
    bg_attr :path

    public
    def to_s
        
        res = @path ? "#{@path.format}: " : ""
        res << "Can't cast #@actual to #@expected"
    end
end

module CastImpl
 
    module_function

    def fail_cast( val, typ, path )
        
        raise TypeCastError.new(
            :actual => Mingle.type_of( val ),
            :expected => typ,
            :path => path
        )
    end

    def cast_num( val, typ, path )
        
        cls, meth = 
            case typ
            when TYPE_INT32 then [ MingleInt32, :to_i ]
            when TYPE_INT64 then [ MingleInt64, :to_i ]
            when TYPE_UINT32 then [ MingleUint32, :to_i ]
            when TYPE_UINT64 then [ MingleUint64, :to_i ]
            when TYPE_FLOAT32 then [ MingleFloat32, :to_f ]
            when TYPE_FLOAT64 then [ MingleFloat64, :to_f ]
            else raise "Bad num type: #{typ}"
            end
        
        case val
        when cls then val
        when MingleString then cls.new( val.to_s.send( meth ) )
        when MingleNumber then cls.new( val.num.send( meth ) )
        else fail_cast( val, typ, path )
        end
    end

    def cast_timestamp( val, typ, path )
        
        case val
        when MingleTimestamp then val
        when MingleString then MingleTimestamp.rfc3339( val )
        else fail_cast( val, typ, path )
        end
    end

    def cast_string( val, typ, path )
        
        case val
        when MingleString then val
        when MingleTimestamp then val.rfc3339
        when MingleNumber, MingleBoolean then MingleString.new( val.to_s )
        else fail_cast( val, typ, path )
        end
    end

    def cast_atomic_value( val, typ, path )
        
        case typ
        when TYPE_INT32, TYPE_INT64, TYPE_UINT32, TYPE_UINT64, 
             TYPE_FLOAT32, TYPE_FLOAT64 
            cast_num( val, typ, path )
        when TYPE_TIMESTAMP then cast_timestamp( val, typ, path )
        when TYPE_STRING then cast_string( val, typ, path )
        else raise "Can't cast to #{typ}"
        end
    end

    def cast_value( val, typ, path )

        case typ
        when AtomicTypeReference then cast_atomic_value( val, typ, path )
        else raise "Unimplemented"
        end
    end
end

def self.cast_value( val, typ, path = nil )
    CastImpl.cast_value( val, typ, path )
end

def self.quote_value( val )
    case val
    when MingleString then Chars.external_form_of( val.to_s )
    when MingleInt32, MingleInt64, MingleUint32, MingleUint64, 
         MingleFloat32, MingleFloat64 
        val.to_s
    when MingleTimestamp then Chars.external_form_of( val.rfc3339 )
    else raise "Can't quote: #{val} (#{val.class})"
    end
end

module IoConstants

    TYPE_CODE_NIL = 0x00
    TYPE_CODE_ID = 0x01
    TYPE_CODE_NS = 0x02
    TYPE_CODE_DECL_NM = 0x03
    TYPE_CODE_QN = 0x04
    TYPE_CODE_ATOM_TYP = 0x05
    TYPE_CODE_LIST_TYP = 0x06
    TYPE_CODE_NULLABLE_TYP = 0x07
    TYPE_CODE_REGEX_RESTRICT = 0x08
    TYPE_CODE_RANGE_RESTRICT = 0x09
    TYPE_CODE_BOOL = 0x0a
    TYPE_CODE_STRING = 0x0b
    TYPE_CODE_INT32 = 0x0c
    TYPE_CODE_INT64 = 0x0d
    TYPE_CODE_FLOAT32 = 0x0e
    TYPE_CODE_FLOAT64 = 0x0f
    TYPE_CODE_TIME_RFC3339 = 0x10
    TYPE_CODE_BUFFER = 0x11
    TYPE_CODE_ENUM = 0x12
    TYPE_CODE_SYM_MAP = 0x13
    TYPE_CODE_MAP_PAIR = 0x14
    TYPE_CODE_STRUCT = 0x15
    TYPE_CODE_LIST = 0x17
    TYPE_CODE_END = 0x18
end

class BinIoError < StandardError; end

class BinIoBase < BitGirderClass
 
    include IoConstants

    private
    def error( msg )
        BinIoError.new( msg )
    end

    private
    def errorf( msg, *argv )
        error( sprintf( msg, *argv ) )
    end
end

class BinReader < BinIoBase

    private_class_method :new

    bg_attr :rd # A BitGirder::Io::BinaryReader

    private
    def peek_type_code
        @rd.peek_int8
    end

    private
    def read_type_code
        @rd.read_int8
    end

    private
    def expect_type_code( tc )
        
        if ( act = read_type_code ) == tc
            tc
        else
            raise errorf( "Expected type code 0x%02x but got 0x%02x", tc, act )
        end
    end

    private
    def buf32_as_utf8

        RubyVersions.when_19x( @rd.read_buffer32 ) do |buf|
            buf.force_encoding( "utf-8" )
        end
    end

    public
    def read_identifier
        
        expect_type_code( TYPE_CODE_ID )

        parts = Array.new( @rd.read_uint8 ) { buf32_as_utf8 }
        MingleIdentifier.send( :new, :parts => parts )
    end

    private
    def read_identifiers
        Array.new( @rd.read_uint8 ) { read_identifier }
    end

    public
    def read_namespace
        
        expect_type_code( TYPE_CODE_NS )

        MingleNamespace.send( :new,
            :parts => read_identifiers,
            :version => read_identifier
        )
    end

    private
    def read_declared_type_name
    
        expect_type_code( TYPE_CODE_DECL_NM )

        DeclaredTypeName.send( :new, :name => buf32_as_utf8 )
    end

    public
    def read_qualified_type_name
        
        expect_type_code( TYPE_CODE_QN )

        QualifiedTypeName.new(
            :namespace => read_namespace,
            :name => read_declared_type_name
        )
    end

    public
    def read_type_name

        case tc = peek_type_code
        when TYPE_CODE_DECL_NM then read_declared_type_name
        when TYPE_CODE_QN then read_qualified_type_name
        else raise errorf( "Unrecognized type name code: 0x%02x", tc )
        end
    end

    private
    def read_restriction
        
        if ( tc = read_type_code ) == TYPE_CODE_NIL
            nil
        else
            raise error( "Non-nil restrictions not yet implemented" )
        end
    end

    private
    def read_atomic_type_reference
        
        expect_type_code( TYPE_CODE_ATOM_TYP )

        AtomicTypeReference.send( :new,
            :name => read_type_name,
            :restriction => read_restriction
        )
    end

    public
    def read_type_reference

        case tc = peek_type_code
        when TYPE_CODE_ATOM_TYP then read_atomic_type_reference
        when TYPE_CODE_LIST_TYP then read_list_type_reference
        when TYPE_CODE_NULLABLE_TYP then read_nullable_type_reference
        else raise errorf( "Unrecognized type reference code: 0x%02x", tc )
        end
    end

    def self.as_bin_reader( io_rd )
        self.send( :new, :rd => io_rd )
    end
end

class BinWriter < BinIoBase
    
    bg_attr :wr

    private_class_method :new

    private
    def write_type_code( tc )
        @wr.write_uint8( tc )
    end

    private
    def write_nil
        write_type_code( TYPE_CODE_NIL )
    end

    private
    def write_qualified_type_name( qn )
        
        write_type_code( TYPE_CODE_QN )
        write_namespace( qn.namespace )
        write_declared_type_name( qn.name )
    end

    public
    def write_identifier( id )
        
        write_type_code( TYPE_CODE_ID )

        @wr.write_uint8( id.parts.size )
        id.parts.each { |part| @wr.write_buffer32( part ) }
    end

    private
    def write_identifiers( ids )

        @wr.write_uint8( ids.size )
        ids.each { |id| write_identifier( id ) }
    end

    private
    def write_namespace( ns )

        write_type_code( TYPE_CODE_NS )
        write_identifiers( ns.parts )
        write_identifier( ns.version )
    end
    
    private
    def write_declared_type_name( nm )
        
        write_type_code( TYPE_CODE_DECL_NM )
        @wr.write_buffer32( nm.name )
    end

    private
    def write_type_name( nm )

        case nm
        when DeclaredTypeName then write_declared_type_name( nm )
        when QualifiedTypeName then write_qualified_type_name( nm )
        else raise error( "Unhandled type name: #{nm.class}" )
        end
    end

    private
    def write_atomic_type_reference( typ )
        
        write_type_code( TYPE_CODE_ATOM_TYP )
        write_type_name( typ.name )

        case typ.restriction
        when nil then write_nil
        else raise error( "Unhandled restriction: #{typ}" )
        end
    end

    public
    def write_type_reference( typ )
        
        case typ
        when AtomicTypeReference then write_atomic_type_reference( typ )
        when ListTypeReference then write_list_type_reference( typ )
        when NullableTypeReference then write_nullable_type_reference( typ )
        else raise error( "Unhandled type reference: #{typ.class}" )
        end
    end

    def self.as_bin_writer( io_wr )
        self.send( :new, :wr => io_wr )
    end
end

class MingleServiceRequest < BitGirder::Core::BitGirderClass
    
    bg_attr :namespace, :processor => MingleNamespace

    bg_attr :service, :processor => MingleIdentifier

    bg_attr :operation, :processor => MingleIdentifier
    
    bg_attr :authentication,
            :required => false,
            :processor => lambda { |v| MingleModels.as_mingle_value( v ) }

    bg_attr :parameters, 
            :default => MingleSymbolMap::EMPTY,
            :processor => MingleSymbolMap
end

class MingleServiceResponse < BitGirder::Core::BitGirderClass
    
    def initialize( res, ex )

        @res = res
        @ex = ex
    end

    public
    def ok?
        ! @ex
    end

    alias is_ok ok?

    public
    def get

        if ok?
            @res
        else
            MingleModels.raise_as_ruby_error( @ex )
        end
    end

    public
    def get_result

        if ok?
            @res
        else
            raise "get_res called on non-ok response"
        end
    end

    alias result get_result

    public
    def get_error
        
        if ok?
            raise "get_error called on ok response"
        else
            @ex
        end
    end

    alias error get_error

    public
    def to_s
        super.inspect
    end

    def self.create_success( res )

        res = MingleModels.as_mingle_value( res );
        res = nil if res.is_a?( MingleNull )

        MingleServiceResponse.new( res, nil )
    end

    def self.create_failure( ex )
        MingleServiceResponse.new( nil, ex )
    end
end

end
