require 'bitgirder/core'

module BitGirder
module Io
 
require 'fileutils'
require 'tempfile'
require 'json'
require 'yaml'
require 'base64'
require 'tmpdir'

include BitGirder::Core
include BitGirderMethods
extend BitGirderMethods

ORDER_LITTLE_ENDIAN = :little_endian
ORDER_BIG_ENDIAN = :big_endian

# Lazily load and assert presence of utf8 encoding
def enc_utf8

    @@enc_utf8 ||= 
       ( Encoding.find( "utf-8" ) or raise "No utf-8 encoding found (?!)" )
end

module_function :enc_utf8

def fu()
    BitGirderLogger.get_logger.is_debug? ? FileUtils::Verbose : FileUtils
end

module_function :fu

def as_encoded( str, enc )
    
    not_nil( str, :str )
    not_nil( enc, :enc )

    str.encoding == enc ? str : str.encode( enc )
end

module_function :as_encoded

def strict_encode64( str )

    if RUBY_VERSION >= "1.9"
        Base64.strict_encode64( str )
    else
        Base64.encode64( str ).split( /\n/ ).join( "" )
    end
end

module_function :strict_encode64

# Despite the name, this method only enforces strictness when running in a ruby
# >= 1.9 for now, though we may backport and handcode integrity checks at some
# point for other rubies
def strict_decode64( str )
    
    if RUBY_VERSION >= "1.9"
        Base64.strict_decode64( str )
    else
        Base64.decode64( str )
    end
end

module_function :strict_decode64

# Returns i as a little-endian 2's complement byte array; Algorithm is from
# http://stackoverflow.com/questions/5284369/ruby-return-byte-array-containing-twos-complement-representation-of-bignum-fi
# (with some cosmetic differences, including the return val's endianness).
#
# Though not stated there, it's worth noting that the reason for the end
# condition test of the 7th (sign) bit is to avoid stopping prematurely on
# inputs such as 0xff00, dropping the sign information from the result
#
def int_to_byte_array( i )
    
    not_nil( i, :i )

    res = []

    begin
        res << ( i & 0xff )
        i >>= 8
    end until ( i == 0 || i == -1 ) && ( res[ -1 ][ 7 ] == i[ 7 ] )

    res
end

module_function :int_to_byte_array

def file_exists( d )
    
    raise "File or directory #{d} does not exist" unless File.exist?( d )
    d
end

module_function :file_exists

def open_tempfile( basename = nil, *rest, &blk )
    
    basename ||= [ "bg-tmp-", ".tmp" ]
    Tempfile::open( basename, *rest, &blk )
end

def mktmpdir( *argv, &blk )
    Dir.mktmpdir( *argv, &blk )
end

module_function :mktmpdir

def fsize( obj )
    
    stat =
        case obj
        when File, Tempfile then File.stat( obj.path )
        when String then File.stat( obj )
        else raise TypeError, "Unhandled type for fsize: #{obj.class}"
        end
    
    stat.size
end

module_function :fsize

module_function :open_tempfile

def write_file( text, file )
 
    fu().mkdir_p( File.dirname( file ) )
    File.open( file, "w" ) { |io| io.print text }
end

module_function :write_file

def first_line( file )
    File.open( file ) { |io| io.readline.chomp }
end

module_function :first_line

def slurp_io( io, blk_sz = 4096 )
    
    not_nil( io, :io )

    while s = io.read( blk_sz ); res = res ? res << s : s; end
    
    res
end

module_function :slurp_io

def slurp( io, blk_sz = 4096 )
    
    case io
        when IO, Tempfile then slurp_io( io, blk_sz )
        when String then File.open( io ) { |io2| slurp_io( io2, blk_sz ) }
        else raise "Unknown slurp target: #{io} (#{io.class})"
    end
end

module_function :slurp

def as_write_dest( obj )
    
    not_nil( obj, :obj )

    case obj
        when IO, Tempfile then yield( obj )
        when String then File.open( obj, "w" ) { |io| yield( io ) }
        else raise TypeError, "Unknown write dest: #{obj.class}"
    end
end

module_function :as_write_dest

def as_read_src( obj )
    
    not_nil( obj, :obj )

    case obj
        when IO, Tempfile then yield( obj )
        when String then File.open( file_exists( obj ) ) { |io| yield( io ) }
        else raise TypeError, "Unkown read src: #{obj.class}"
    end
end

module_function :as_read_src

def as_json( obj ); JSON.generate( not_nil( obj, :obj ) ); end

module_function :as_json

def parse_json( str ); JSON.parse( not_nil( str, :str ) ); end

module_function :parse_json

def dump_json( obj, file )
 
    not_nil( obj, :obj )
    not_nil( file, :file )

    json = as_json( obj )

    as_write_dest( file ) { |io| io.print( json ) }
end

module_function :dump_json

def load_json( file )
 
    not_nil( file, :file )
    as_read_src( file ) { |io| parse_json( slurp( io ) ) }
end

module_function :load_json

def dump_yaml( obj, dest )
    
    not_nil( obj, :obj )
    not_nil( dest, :dest )

    as_write_dest( dest ) { |io| YAML.dump( obj, io ) }
end

module_function :dump_yaml

def load_yaml( src )
    
    not_nil( src, :src )

    as_read_src( src ) { |io| YAML.load( io ) }
end

module_function :load_yaml

def is_executable( file )
 
    file_exists( file )
    raise "Not an executable: #{file}" unless File.executable?( file )

    file
end

module_function :is_executable

# Effectively enables mdkir_p as an inline function to enable statements
# like:
# 
#   some_dir = ensure_dir( some_dir )
#
def ensure_dir( d )

    fu().mkdir_p( d ) unless File.exist?( d )
    d
end

module_function :ensure_dir

def ensure_wiped( d )
    
    fu().rm_rf( d )
    ensure_dir( d )
end

module_function :ensure_wiped

def ensure_dirs( *dirs )
    dirs.each { |dir| ensure_dir( dir ) }
end

module_function :ensure_dirs

# Ensures that the directory referred to as dirname( file ) exists, and
# returns file itself (not the parent). Fails if dirname does not return
# anything meaningful for file.
def ensure_parent( file )
    
    parent = File.dirname( file )
    raise "No parent exists for #{file}" unless parent
    
    ensure_dir( parent )

    file
end

module_function :ensure_parent

def which( cmd, fail_on_miss = false )
    
    not_nil( cmd, "cmd" )

    if ( f = `which #{cmd}`.strip ) and f.empty? 
        raise "Cannot find command #{cmd.inspect} in path" if fail_on_miss
    else
        f
    end
end

module_function :which

def debug_wait2( opts )
    
    not_nil( opts, "opts" )
    
    pid = has_key( opts, :pid )

    name = opts[ :name ]
    name_str = name ? "#{name} (pid #{pid})" : pid.to_s
    code( "Waiting on #{name_str}" )

    pid, status = Process::wait2( pid )
    msg = "Process #{pid} exited with status #{status.exitstatus}"

    if status.success?
        code( msg )
    else
        opts[ :check_status ] ? ( raise msg ) : warn( msg )
    end

    [ pid, status ]
end

module_function :debug_wait2

def debug_kill( sig, pid, opts = {} )
    
    not_nil( opts, "opts" )
    
    msg = "Sending #{sig} to #{pid}"
    msg += " (#{opts[ :name ]})" if opts.key?( :name )

    BitGirderLogger.get_logger.code( msg )
    Process.kill( sig, pid )
end

module_function :debug_kill

def read_full( io, len, buf = nil )
    
    args = [ len ]
    args << buf if buf
    buf = io.read( *args )

    if ( sz = buf == nil ? 0 : buf.bytesize ) < len
        raise EOFError.new( "EOF after #{sz} bytes (wanted #{len})" )
    end

    buf
end

module_function :read_full

class DataUnitError < StandardError; end

class DataUnit < BitGirderClass
    
    bg_attr :name
    bg_attr :byte_scale

    private_class_method :new
    
    BYTE = self.send( :new, :name => :byte, :byte_scale => 1 )
    KILOBYTE = self.send( :new, :name => :kilobyte, :byte_scale => 2 ** 10 )
    MEGABYTE = self.send( :new, :name => :megabyte, :byte_scale => 2 ** 20 )
    GIGABYTE = self.send( :new, :name => :gigabyte, :byte_scale => 2 ** 30 )
    TERABYTE = self.send( :new, :name => :terabyte, :byte_scale => 2 ** 40 )
    PETABYTE = self.send( :new, :name => :petabyte, :byte_scale => 2 ** 50 )

    UNITS = [ BYTE, KILOBYTE, MEGABYTE, GIGABYTE, TERABYTE, PETABYTE ]

    map_instance_of( String, Symbol ) do |val|
        
        # Avoid string ops on val if we can
        if res = UNITS.find { |u| u.name == res }
            return res
        end
        
        case val.to_s.downcase.to_sym
        when :b, :byte, :bytes then BYTE
        when :k, :kb, :kilobyte, :kilobytes then KILOBYTE
        when :m, :mb, :megabyte, :megabytes then MEGABYTE
        when :g, :gb, :gigabyte, :gigabytes then GIGABYTE
        when :t, :tb, :terabyte, :terabytes then TERABYTE
        when :p, :pb, :petabyte, :petabytes then PETABYTE
        else raise DataUnitError, "Unknown data unit: #{val}"
        end
    end
end

class DataSizeError < StandardError; end

class DataSize < BitGirderClass

    include Comparable

    bg_attr :unit, :processor => DataUnit
    
    bg_attr :size, :validation => :nonnegative

    map_instance_of( Integer ) do |val| 
        DataSize.new( :size => val, :unit => DataUnit::BYTE )
    end

    map_instance_of( String ) do |val|

        if md = /^\s*(\d+)\s*([a-zA-Z]*)\s*$/.match( val )
            if ( unit = md[ 2 ] ) == "" then unit = DataUnit::BYTE end
            DataSize.new( :size => md[ 1 ].to_i, :unit => unit )
        else
            raise DataSizeError, "Invalid data size: #{val.inspect}"
        end
    end

    public
    def bytes
        @size * @unit.byte_scale
    end

    public
    def ==( o )
        return true if o.equal?( self )
        return false unless o.is_a?( DataSize )
        return self.bytes == o.bytes
    end

    public
    def <=>( o )
        return nil unless o.is_a?( DataSize )
        return self.bytes <=> o.bytes
    end

    public
    def to_s
        bytes.to_s
    end
end

class BinaryConverter < BitGirderClass

    require 'bigdecimal'

    bg_attr :order,
            :processor => :symbol,
            :validation => lambda { |o| 
                unless o == ORDER_LITTLE_ENDIAN || o == ORDER_BIG_ENDIAN
                    raise "Unrecognized order: #{o}"
                end
            }
        
    # Set @plat_order as an instance variable instead of as a class constant. We
    # could theoretically set plat order as a class constant, but don't for 2
    # reasons. One is that, in the event there is a bug in our detection, we
    # don't want to fail any class requiring this file, since some code may
    # never touch the BinaryConverter class anyway. Two is that we can use
    # reflection during testing to simulate a different platform byte ordering
    # on a test-by-test basis, rather than having to change a global constant
    def initialize( *argv )
        
        super( *argv )

        @plat_order = BinaryConverter.detect_platform_order
    end

    private
    def is_plat_order?
        @order == @plat_order
    end

    private
    def impl_write_int( num, nm, fmt )
        
        not_nil( num, nm )
        res = [ num ].pack( fmt )
        
        res.reverse! unless is_plat_order?

        res
    end

    private
    def impl_read_int( str, nm, fmt )

        not_nil( str, nm )

        str = str.reverse unless is_plat_order?
        str.unpack( fmt )[ 0 ]
    end

    {
        :int8 => "c",
        :int32 => "l",
        :int64 => "q"
    }.
    each_pair do |type, fmt|
        
        [ [ type, fmt ], [ :"u#{type}", fmt.upcase ] ].each do |pair|
        
            type2, fmt2 = *pair
            wm, rm = %w{ write read }.map { |s| :"#{s}_#{type2}" }
    
            define_method( wm ) { |i| impl_write_int( i, :i, fmt2 ) }
            define_method( rm ) { |s| impl_read_int( s, :s, fmt2 ) }
    
            public wm, rm
        end
    end

    private
    def get_float_format( is64 )
        
        fmt = 
            case @order
                when ORDER_LITTLE_ENDIAN then 'e'
                when ORDER_BIG_ENDIAN then 'g'
                else raise "Unhandled order: #@order"
            end
        
        fmt = fmt.upcase if is64

        fmt
    end

    private
    def impl_write_float( num, nm, is64 )
        
        not_nil( num, nm )
        
        fmt = get_float_format( is64 )
        [ num ].pack( fmt )
    end

    public
    def write_float32( f )
        impl_write_float( f, :f, false )
    end

    public
    def write_float64( d )
        impl_write_float( d, :d, true )
    end

    private
    def impl_read_float( str, nm, is64 )
        
        not_nil( str, nm )

        fmt = get_float_format( is64 )
        str.unpack( fmt )[ 0 ]
    end

    public
    def read_float32( s )
        impl_read_float( s, :s, false )
    end

    public
    def read_float64( s )
        impl_read_float( s, :s, true )
    end

    public
    def write_bignum( i )
        
        not_nil( i, :i )

        arr = Io.int_to_byte_array( i )
        arr.reverse! if @order == ORDER_BIG_ENDIAN

        res = write_int32( arr.size )
        arr.each { |b| res << b.chr }

        res
    end

    private
    def inflate_bignum( s )
 
        # Ensure we process bytes in big-endian order
        s = s.reverse if @order == ORDER_LITTLE_ENDIAN

        res = s[ 0, 1 ].unpack( 'c' )[ 0 ] # res is now set and correctly signed

        s[ 1 .. -1 ].each_byte do |b|
            res <<= 8
            res += b
        end

        res
    end

    public
    def read_bignum( s )
        read_bignum_with_info( s )[ 0 ]
    end

    public
    def read_bignum_with_info( s )

        not_nil( s, :s )
        
        if ( len = s.size ) < 4
            raise "Input does not have a size (input len: #{len})"
        elsif len == 4
            raise "Input has no integer data (input len: #{len})"
        end

        sz = read_int32( s[ 0, 4 ] )
        rem = [ len - 4, len - sz ].min

        if ( rem = s.size - 4 ) < sz
            raise "Input specifies size #{sz} but actually is only of " \
                   "length #{rem}"
        end

        info = { :total_len => 4 + sz, :hdr_len => 4, :num_len => sz }
        [ inflate_bignum( s[ 4, sz ] ), info ]
    end

    # write_bigdec (read_bigdec does the inverse) interprets ruby BigDecimal
    # first as an unscaled value and a scale as defined in Java's BigDecimal
    # class, and then serializes them as the concatenation of write_bignum(
    # unscaled ) and write_int32( scale ), with each component's byte-order
    # determined by @order
    public
    def write_bigdec( n )
 
        not_nil( n, :n )
        raise "Not a BigDecimal" unless n.is_a?( BigDecimal )

        sign, unscaled, base, exp = n.split
        raise "Unexpected base: #{base}" unless base == 10
        raise "NaN not supported" if sign == 0

        if ( unscaled_i = unscaled.to_i ) == 0
            write_bignum( 0 ) + write_int32( 0 )
        else
            # sci_exp is the exponent we'd get using scientific notation
            sci_exp = -unscaled.size + exp
            write_bignum( unscaled_i * sign ) + write_int32( sci_exp )
        end
    end

    public
    def read_bigdec_with_info( s )
        
        not_nil( s, :s )

        unscaled, info1 = read_bignum_with_info( s )
        unscaled_len = has_key( info1, :total_len )

        scale = read_int32( s[ unscaled_len, 4 ] )

        info2 = { 
            :total_len => unscaled_len + 4, 
            :unscaled_len => unscaled_len, 
            :scale_len => 4
        }

        num_str = unscaled.to_s( 10 ) + "e" + ( scale ).to_s

        [ BigDecimal.new( num_str ), info2 ]
    end

    public
    def read_bigdec( s )
        read_bigdec_with_info( s )[ 0 ]
    end

    def self.detect_platform_order
    
        case test = [ 1 ].pack( 's' )
            when "\x01\x00" then ORDER_LITTLE_ENDIAN
            when "\x00\x01" then ORDER_BIG_ENDIAN
            else raise "Undetected byte order: #{test.inspect}"
        end
    end
end

class BinaryIo < BitGirderClass
    
    bg_attr :io
    bg_attr :order

    attr_reader :conv

    # pos returns the zero-indexed position of the next byte that would be read,
    # in the case of a reader, or the number of bytes that have been written in
    # the case of a writer.  This value is only valid as long as no exceptions
    # have been encountered in any of the read|write methods and all access to
    # the underlying io object has been through this instance the read* methods 
    attr_reader :pos

    public
    def close
        @io.close if @io.respond_to?( :close )
    end

    private
    def impl_initialize

        super
        @conv = BinaryConverter.new( :order => @order )
        @pos = 0
    end

    def self.new_with_order( ord, opts )
        self.new( { :order => ord }.merge( opts ) )
    end

    def self.new_le( opts )
        self.new_with_order( ORDER_LITTLE_ENDIAN, opts )
    end

    def self.new_be( opts )
        self.new_with_order( ORDER_BIG_ENDIAN, opts )
    end
end

class BinaryWriter < BinaryIo
    
    {
        :int8 => 1,
        :int32 => 4,
        :int64 => 8,
        :uint8 => 1,
        :uint32 => 4,
        :uint64 => 8,
        :float32 => 4,
        :float64 => 8

    }.each_pair do |meth, sz|

        meth = :"write_#{meth}"

        define_method( meth ) do |val| 
            @io.write( @conv.send( meth, val ) )
            @pos += sz
        end
    end

    public
    def write_bool( b )
        write_int8( b ? 1 : 0 )
    end

    public
    def write_full( buf )
        @io.write( buf )
        @pos += buf.bytesize
    end

    alias write write_full

    private
    def impl_write_buffer32( buf, sz )
        
        write_int32( sz )
        write_full( buf )
    end

    public
    def write_buffer32( buf )

        not_nil( buf, :buf )
        impl_write_buffer32( buf, buf.bytesize )
    end

    public
    def write_utf8( str )

        not_nil( str, :str )
        
        RubyVersions.when_19x { str = Io.as_encoded( str, Encoding::UTF_8 ) }

        impl_write_buffer32( str, str.bytesize )
    end
end

class BinaryReader < BinaryIo

    public
    def eof?
        @io.eof?
    end

    public
    def peekc
        
        res = @io.getc.tap { |c| @io.ungetc( c ) }

        case res
        when String then res
        when Fixnum then res.chr
        else raise "Unexpected getc val: #{res.class}"
        end
    end

    public
    def peek_int8
        peekc.unpack( 'c' )[ 0 ]
    end

    public
    def read_full( len, buf = nil )
        Io.read_full( @io, len, buf ).tap { @pos += len }
    end

    public
    def read( *argv )
        @io.read( *argv ).tap { |res| @pos += res.bytesize }
    end
 
    { 
        :int8 => 1, 
        :int32 => 4, 
        :int64 => 8, 
        :uint8 => 1, 
        :uint32 => 4, 
        :uint64 => 8, 
        :float32 => 4, 
        :float64 => 8 

    }.each_pair do |k, v|
        meth = :"read_#{k}"
        define_method( meth ) { @conv.send( meth, read_full( v ) ) }
    end

    public
    def read_bool
        read_int8 != 0
    end

    alias read_boolean read_bool

    public
    def read_buffer32
        read_full( read_int32 )
    end

    public
    def read_utf8
 
        len = read_int32

        str = "" * len
        read_full( len, str )
        RubyVersions.when_19x { str.force_encoding( "utf-8" ) }

        str
    end
end

class UnixProcessBuilder < BitGirderClass
 
    bg_attr :cmd
 
    bg_attr :argv, :default => []

    bg_attr :env, :default => {}

    bg_attr :opts, :default => {}

    bg_attr :show_env_in_debug, :mutable => true, :required => false

    def initialize( opts )
        super( opts )
    end

    private
    def debug_call( opts )

        dbg = { :cmd => @cmd, :argv => @argv, :opts => @opts }
        dbg[ :env ] = @env if @show_env_in_debug

        code( "Doing #{opts[ :call_type ]} with #{dbg.inspect}" )
    end

    private
    def str_map( val )
        
        case val

            when Array then val.map { |o| str_map( o ) }

            when Hash 
                val.inject( {} ) do |h, pair|
                    h[ str_map( pair[ 0 ] ) ] = str_map( pair[ 1 ] )
                    h
                end
            
            else val.to_s
        end
    end

    private
    def get_call_argv
        
        [ str_map( @env ), str_map( @cmd ), str_map( @argv ), @opts ].flatten
    end

    def get_call_argv18
        
        res = get_call_argv
        [ res[ 1 ], *( res[ 2 ] ) ] # [ cmd, *argv ]
    end

    public
    def spawn

        debug_call( :call_type => "spawn" )
        Process.spawn( *get_call_argv )
    end

    public
    def exec

        debug_call( :call_type => "exec" )
        Kernel.exec( *get_call_argv )
    end
 
    public
    def system( opts = {} )
 
        debug_call( :call_type => "system" )
        
        proc_res_raise = lambda {
            raise "Command exited with status #{$?.exitstatus}"
        }

        if RUBY_VERSION >= "1.9"
            proc_res_raise.call unless Kernel.system( *get_call_argv )
        else
            if pid = Kernel.fork
                Process.wait( pid )
                proc_res_raise.call unless $?.success?
            else
                Kernel.exec( *get_call_argv18 )
            end
        end
    end

    public
    def popen( mode, &blk )
        
        debug_call( :call_type => "popen" )

        if RUBY_VERSION >= "1.9"
            IO.popen( get_call_argv, mode, &blk )
        else
            IO.popen( get_call_argv18.join( " " ), mode, &blk )
        end
    end

    [ :spawn, :exec, :system, :popen ].each do |meth|
        self.class.send( :define_method, meth ) { |opts| 
            self.new( opts ).send( meth )
        }
    end
end

end
end
