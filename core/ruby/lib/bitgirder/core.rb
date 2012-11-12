# Placed here so that the rest of the libs and executables in our codebase
# needn't worry about requiring rubygems when being run in ruby 1.8.
require 'rubygems' if RUBY_VERSION < "1.9"

module BitGirder
module Core

ENV_BITGIRDER_DEBUG = "BITGIRDER_DEBUG"

EXIT_SUCCESS = 0
EXIT_FAILURE = 1

class RubyVersions

    def self.when_geq( ver, val = nil )
        if RUBY_VERSION >= ver
            yield( val )
        else
            val
        end
    end

    def self.when_19x( val = nil, &blk )
        self.when_geq( "1.9", val, &blk )
    end

    def self.jruby?
        RUBY_PLATFORM == "java"
    end
end

class BitGirderLogger

    require 'time'
    require 'thread'
    
    private_class_method :new

    # Our log levels
    CODE = :CODE
    CONSOLE = :CONSOLE
    WARN = :WARN
    DEFAULT = CONSOLE

    attr_reader :level

    def initialize
        @lock = Mutex.new
    end

    # makes a string msg from argv, which can either be an exception and a
    # message or just a message
    private
    def make_msg( argv )
        
        case len = argv.length

            when 1 then argv.shift

            when 2
                e, msg = *argv
                msg << "\n" << e.message.to_s

                bt = e.backtrace
                ( msg << "\n" << bt.join( "\n" ) ) if bt

                msg

            else raise ArgumentError, "Wrong number of arguments: #{len}"
        end
    end

    def self.level_of( lev )
        case lev
            when CODE then 1
            when CONSOLE then 2
            when WARN then 3
            else raise "Invalid level: #{lev}"
        end
    end

    private
    def send_msg( lev, time, argv )
        
        if self.class.level_of( lev ) >= self.class.level_of( @level )
            str = make_msg( argv )
            @lock.synchronize { puts "[#{time.iso8601( 6 )}]: #{lev}: #{str}" }
        end
    end

    public
    def code( *argv )
        send_msg( CODE, Time.now, argv )
    end

    alias debug code

    public
    def warn( *argv )
        send_msg( WARN, Time.now, argv )
    end

    public
    def console( *argv )
        send_msg( CONSOLE, Time.now, argv )
    end

    public
    def is_debug?
        @level == CODE
    end

    public
    def level=( lev )
 
        @level =
            case lev
                when CODE, CONSOLE, WARN then lev
                else raise ArgumentError, "Unknown level: #{lev}"
            end
    end

    def self.is_debug_env_set?
        ( ENV[ ENV_BITGIRDER_DEBUG ] or "" ).strip =~ /^(true|yes)$/
    end 

    @logger = new
    @logger.level = self.is_debug_env_set? ? CODE : DEFAULT

    def self.get_logger
        @logger
    end

    class <<self
    
        [ :debug, :code, :warn, :console ].each do |meth|
    
            define_method( meth ) do |*argv|
                get_logger.send( meth, *argv )
            end
        end
    end
end

module Reflect
    
    # Returns an array of Symbol regardless of ruby version (1.8 uses String,
    # 1.9 Symbols)
    def self.instance_methods_of( v )

        res = v.instance_methods

        if res.size > 0 && res[ 0 ].is_a?( String )
            res = res.map { |s| s.to_sym } 
        end

        res
    end
end

module BitGirderMethods

    module_function

    def code( *argv )
        BitGirderLogger.get_logger.code( *argv )
    end

    def console( *argv )
        BitGirderLogger.get_logger.console( *argv )
    end

    alias debug code

    def warn( *argv )
        BitGirderLogger.get_logger.warn( *argv )
    end

    # Each constant corresponds to a prefix in an error message appropriate to
    # that type
    #
    PARAM_TYPE_ARG = "Argument"
    PARAM_TYPE_ENVVAR = "Environment Variable"
    PARAM_TYPE_KEY = "Value for key"

    def check_fail_prefix( name, param_type )

        # Since we'll most often be using symbols instead of strings for arg
        # parameter names we special case displaying them as strings
        name_str = 
            if param_type == PARAM_TYPE_ARG && name.is_a?( Symbol )
                %{"#{name_str}"}
            else
                name.inspect
            end

        "#{param_type} #{name.inspect}"
    end

    def not_nil( val, name = nil, param_type = PARAM_TYPE_ARG )
        
        if val.nil?
            if name
                raise "#{check_fail_prefix( name, param_type )} cannot be nil"
            else
                raise "Value is nil"
            end
        else
            val
        end
    end

    def compares_to( val, 
                     comp_target, 
                     comp_type, 
                     name = nil, 
                     param_type = PARAM_TYPE_ARG )
        
        if val.send( comp_type, comp_target )
            val
        else
            raise "#{check_fail_prefix( name, param_type )}' is out of range " \
                  "(#{comp_target} #{comp_type} #{val})"
        end
    end

    def nonnegative( val, name = nil, param_type = PARAM_TYPE_ARG )
        compares_to( val, 0, :>=, name, param_type )
    end

    def positive( val, name = nil, param_type = PARAM_TYPE_ARG )
        compares_to( val, 0, :>, name, param_type )
    end

    # changes "some-arg" to :some_arg
    def ext_to_sym( str )
        
        not_nil( str, "str" )

        # do str.to_s in case it's some random string-ish thing (like a sym) but
        # not an actual String
        str.to_s.gsub( /-/, "_" ).to_sym
    end

    # Returns two arrays (both possibly empty) of stuff in argv before the first
    # '--', if any, and stuff after, if any, including any subsequent
    # appearances of '--'
    def split_argv( argv )
        
        not_nil( argv, :argv )

        first, last = [], nil

        argv.each do |arg|
            
            if last then last << arg
            else
                if arg == "--" then last = []; else first << arg end
            end
        end

        [ first, last ]
    end

    # changes "some-arg" to SomeArg (we can add in namespace qualifiers too if
    # needed, for instance 'some-mod/some-other-mod/some-class' becomes
    # SomeMod::SomeOtherMod::SomeClass)
    def ext_to_class_name( str )

        not_nil( str, "str" )
    
        str = str.to_s

        res = ""

        if str.size > 0
            res = str[ 0 ].upcase + 
                  str[ 1 .. -1 ].gsub( /-(.)/ ) { |m| m[ 1 ].upcase }
        end

        res
    end

    # Changes SomeClass to :some_class
    def class_name_to_sym( cls_name )
 
        not_nil( cls_name, "cls_name" )

        if cls_name.size == 0
            raise "Name is empty string"
        else
            str = cls_name[ 0 ].downcase +
                  cls_name[ 1 .. -1 ].
                    gsub( /([A-Z])/ ) { |m| "_#{m.downcase}" }

            str.to_sym
        end
    end

    # Changes :some_sym to "some-sym"
    def sym_to_ext_id( sym )
        not_nil( sym, "sym" ).to_s.gsub( /_/, "-" )
    end

    # Changes :some_sym to --some-sym ('cli' == 'command line interface')
    def sym_to_cli_switch( sym )
        "--" + sym_to_ext_id( sym )
    end

    # helper method to do the sym-conversion part only (no validity checking
    # done on either the sym or val)
    def set_var( sym, val )
        instance_variable_set( "@#{sym}".to_sym, val )
    end

    # It is assumed that map responds to []= and is not nil. The 'key' parameter
    # may be nil. Returns the value if key is present and not nil
    #
    def has_key( map, key, param_typ = PARAM_TYPE_KEY )
        not_nil( map[ key ], key, param_typ )
    end

    def has_keys( map, *keys )
        keys.inject( [] ) { |arr, key| arr << has_key( map, key ) }
    end

    def has_env( key )
        has_key( ENV, key, PARAM_TYPE_ENVVAR )
    end

    def set_from_key( map, *syms )
        syms.each { |sym| set_var( sym, has_key( map, sym ) ) }
    end

    def unpack_argv_hash( argh, vars, expct = false )

        vars.inject( [] ) do |arr, sym| 

            val = argh[ sym ]
            
            if val != nil
                arr << val
            else
                raise "Missing parameter: #{sym} in arg hash" if expct 
            end

            arr
        end
    end

    def unpack_argv_array( arr, vars, trace )
        
        if arr.size == vars.size
            arr
        else
            msg = "wrong number of arguments (#{arr.size} for #{vars.size})"
            raise( Exception, msg, trace )
        end
    end

    def argv_to_argh( argv, syms )
        
        argv = unpack_argv_array( argv, syms, caller )

        argh = {}
        argv.size.times { |i| argh[ syms[ i ] ] = argv[ i ] }

        argh
    end

    def to_bool( obj )
        
        case obj
            
            when true then true
            when false then false
            when nil then false
            when /^\s*(true|yes)\s*$/i then true
            when /^\s*(false|no)\s*$/i then false
            else raise "Invalid boolean string: #{obj}"
        end
    end

    def raisef( *argv )
        
        raise if argv.empty?

        cls = nil

        case v = argv.first
            when Exception then raise v
            when Class then cls = argv.shift
        end
        
        argv2 = cls == nil ? [] : [ cls ]
        argv2 << sprintf( *argv ) unless argv.empty?

        raise *argv2
    end
end 

class BitGirderAttribute

    include BitGirderMethods

    class InvalidModifier < RuntimeError; end

    PROCESSOR_BOOLEAN = lambda { |v| BitGirderMethods.to_bool( v ) }

    PROCESSOR_SYMBOL = lambda { |v| v == nil ? nil : v.to_sym }

    # If v is not nil we return calling to_i() on it; if it is nil we return it
    # as-is since nil is used to communicate to the user of the attribute that
    # no value whatsoever was provided (nil.to_i otherwise would return 0)
    PROCESSOR_INTEGER = lambda { |v| v == nil ? nil : v.to_i }

    # See PROCESSOR_INTEGER
    PROCESSOR_FLOAT = lambda { |v| v == nil ? nil : v.to_f }

    VALIDATION_NOT_NIL = lambda { |val| raise "Missing value" if val == nil }
 
    # Implies not nil
    VALIDATION_NOT_EMPTY = lambda { |val| 

        VALIDATION_NOT_NIL.call( val )
        raise "Need at least one" if val.empty? 
    }

    VALIDATION_NONNEGATIVE = lambda { |val| val >= 0 or raise "value < 0" }

    VALIDATION_POSITIVE = lambda { |val| 

        VALIDATION_NOT_NIL.call( val )
        val > 0 or raise "value <= 0" 
    }

    VALIDATION_FILE_EXISTS = lambda { |val|
        
        VALIDATION_NOT_NIL.call( val )

        unless File.exist?( val )
            raise "File or directory #{val} does not exist" 
        end
    }

    VALIDATION_OPT_FILE_EXISTS = lambda { |val|
        val && VALIDATION_FILE_EXISTS.call( val )
    }

    ATTR_MODIFIERS = [
        :identifier, 
        :default, 
        :description, 
        :validation, 
        :processor, 
        :is_list,
        :list_validation,
        :required,
        :mutable
    ]

    attr_reader *ATTR_MODIFIERS

    # Rather than have a class def silently fail on a mis-typed attribute
    # modifier (":processr" instead of ":processor") we explicitly check and
    # fail if any of the supplied mods are not one expected by this class
    private
    def check_valid_modifiers( supplied )

        supplied.each do |key|
            ATTR_MODIFIERS.include?( key ) or 
                raise InvalidModifier.new( 
                    "Invalid attribute modifier: #{key.inspect}" )
        end
    end

    private
    def get_processor( p )
 
        case p
            when :boolean then PROCESSOR_BOOLEAN
            when :symbol then PROCESSOR_SYMBOL
            when :integer then PROCESSOR_INTEGER
            when :float then PROCESSOR_FLOAT

            when Class
                if p.ancestors.include?( BitGirderClass )
                    lambda { |o| p.as_instance( o ) }
                else
                    raise ArgumentError, "Not a #{BitGirderClass}: #{p}"
                end

            else p
        end
    end

    private
    def get_validation( v )

        case v
            when :not_nil then VALIDATION_NOT_NIL
            when :nonnegative then VALIDATION_NONNEGATIVE
            when :positive then VALIDATION_POSITIVE
            when :file_exists then VALIDATION_FILE_EXISTS
            when :not_empty then VALIDATION_NOT_EMPTY
            when :opt_file_exists then VALIDATION_OPT_FILE_EXISTS
            else v
        end
    end

    public
    def initialize( argh )

        check_valid_modifiers( argh.keys )

        set_from_key( argh, :identifier )
        
        [ :default, :description, :validation, 
          :is_list, :list_validation, :required, :mutable ].each do |sym|
            set_var( sym, argh[ sym ] )
        end

        @validation = get_validation( @validation )
        @list_validation = get_validation( @list_validation )
        
        @processor = get_processor( argh[ :processor ] )

        @id_sym = :"@#@identifier"
    end

    public
    def get_instance_value( inst )
        inst.instance_variable_get( @id_sym )
    end
end

class BitGirderClassDefinition

    extend BitGirderMethods
    include BitGirderMethods

    REQUIRED_ATTRS = 
        [ :cls, :attrs, :attr_syms, :decl_order, :instance_mappers ]
    
    attr_reader *REQUIRED_ATTRS

    def initialize( opts )
        
        REQUIRED_ATTRS.each do |attr|

            val = BitGirderMethods.has_key( opts, attr )
            instance_variable_set( :"@#{attr}", val )
        end
    end
    
    public
    def add_attr( opts )

        unless opts.key?( :required ) || opts.key?( :default )
            opts[ :required ] = true
        end

        attr = BitGirderAttribute.new( opts )
 
        if @attrs.key?( attr.identifier )
            raise "Attribute #{attr.identifier.inspect} already defined"
        else
            ident = attr.identifier
            @attrs[ ident ] = attr
            @attr_syms << "@#{ident}".to_sym
            @decl_order << ident
            @cls.send( attr.mutable ? :attr_accessor : :attr_reader, ident )
        end
    end

    private
    def get_supplied_attr_value( hash, ident )
        hash[ ident ] || hash[ ident.to_s ]
    end

    private
    def get_default_val( attr )
        
        case d = attr.default
            when Proc then d.call
            when Class then d.new
            when Array, Hash then d.clone
            else d == nil && attr.is_list ? [] : d
        end
    end

    private
    def apply_processor( attr, val )

        if p = attr.processor
            if attr.is_list
                val = val.map { |elt| p.call( elt ) }
            else
                val = p.call( val )
            end
        else
            val
        end
    end

    private
    def get_initial_value( hash, ident, attr )
        
        # Get the val as the caller-supplied one, which we allow to be either
        # the symbol or string form, supplying the default as needed; note that
        # we use key membership and explicit tests for nil instead of boolean
        # operators since we want to allow the value 'false' from the caller or
        # as a default
        val = hash.key?( ident ) ? hash[ ident ] : hash[ ident.to_s ]

        if val == nil 
            val = get_default_val( attr ) # could still end up being nil
        else 
            val = apply_processor( attr, val )
        end

        val
    end

    private
    def validate_list_attr_val( attr, val )
        
        if list_func = attr.list_validation
            list_func.call( val )
        end

        if func = attr.validation
            val.each { |v| func.call( v ) }
        end
    end

    private
    def validate_scalar_attr_val( attr, val )
 
        if ( func = attr.validation ) && ( attr.required || ( ! val.nil? ) )
            func.call( val )
        end
    end

    private
    def validate_attr_val( attr, val, trace, listener )
        
        begin

            if attr.is_list || attr.list_validation
                validate_list_attr_val( attr, val )
            else
                validate_scalar_attr_val( attr, val )
            end

            val

        rescue Exception => e

            listener.call( attr, e )
            raise e, "#{attr.identifier}: #{e.message}", trace
        end
    end

    private
    def set_attr_value( inst, hash, ident, attr, trace, listener )
 
        val = get_initial_value( hash, ident, attr )

        validate_attr_val( attr, val, trace, listener )

        inst.instance_variable_set( :"@#{attr.identifier}", val )
    end

    private
    def impl_initialize( inst )
        
        if inst.respond_to?( meth = :impl_initialize, true ) 
            inst.send( meth )
        end
    end

    public
    def init_instance( inst, argv )

        argh =
            if argv.size == 0
                {}
            else
                if argv.size == 1 and ( arg = argv[ 0 ] ).is_a?( Hash )
                    arg
                else
                    argv_to_argh( argv, @decl_order )
                end
            end

        trace = caller
        listener = BitGirderClassDefinition.get_validation_listener

        @attrs.each_pair { |ident, attr| 
            set_attr_value( inst, argh, ident, attr, trace, listener ) 
        }

        impl_initialize( inst )
    end

    public
    def hash_instance( inst )
        @decl_order.map { |id| @attrs[ id ].get_instance_value( inst ) }.hash
    end

    @@class_defs = {}
    @@validation_listener = nil

    def self.get_class_defs
        @@class_defs.dup
    end

    def self.get_validation_listener

        if res = @@validation_listener
            @@validation_listener = nil
            res
        else
            lambda { |attr, e| } # no-op is default
        end
    end

    def self.validation_listener=( l )
        
        BitGirderMethods.not_nil( l, :l )
        raise "A validation listener is already set" if @@validation_listener
       
        @@validation_listener = l
    end

    def self.default_instance_mappers_for( cls )
        
        [
            InstanceMapper.new(
                :processor => lambda { |val| val.is_a?( cls ) && val } ),

            InstanceMapper.new(
                :processor => lambda { |val|
                    # Use send in case cls made it private
                    val.is_a?( Hash ) && cls.send( :new, val )
                }
            )
        ]
    end
 
    def self.for_class( cls )

        unless res = @@class_defs[ cls ]
 
            attrs, decl_order =
                if cls == BitGirderClass || cls == BitGirderError
                    [ {}, [] ]
                else 
                    cd = self.for_class( cls.superclass )
                    [ Hash.new.merge( cd.attrs ), Array.new( cd.decl_order ) ]
                end
 
            @@class_defs[ cls ] = res = new(
                :cls => cls,
                :attrs => attrs,
                :attr_syms => attrs.keys.map { |id| "@#{id}".to_sym },
                :decl_order => decl_order,
                :instance_mappers => self.default_instance_mappers_for( cls )
            )
        end
 
        res
    end

    def self.code( *argv )
        BitGirderLogger.code( *argv )
    end

    def self.init_instance( inst, argv )
        self.for_class( inst.class ).init_instance( inst, argv )
    end
end

class InstanceMapper

    include BitGirderMethods
 
    attr_reader :processor

    # Does nil and other checks for public frontends
    def initialize( opts )

        not_nil( opts, :opts )
        @processor = has_key( opts, :processor )
    end
end

module BitGirderStructure

    private
    def impl_initialize; end

    def initialize( *argv )
        BitGirderClassDefinition.init_instance( self, argv )
    end

    public 
    def ==( other )
        
        return true if self.equal?( other )
        return false unless other.class == self.class

        ! self.class.get_class_def.attr_syms.find do |sym|

            other_val = other.instance_variable_get( sym )
            self_val = instance_variable_get( sym )

            other_val != self_val
        end
    end

    alias eql? ==
 
    def self.included( cls )
        
        cls.class_eval do
        
            def self.get_class_def
                BitGirderClassDefinition.for_class( self )        
            end

            def self.bg_attr( id, opts = {} )
        
                argh = { 
                    :identifier => id, 
                    :validation => :not_nil 
                }.
                merge( opts )
 
                self.get_class_def.add_attr( argh )
            end
        
            def self.map_instances( opts )
                
                im = InstanceMapper.new( opts ) # Validation in InstanceMapper
                self.get_class_def.instance_mappers.unshift( im )
            end
        
            def self.map_instance_of( *classes, &blk )
        
                self.map_instances(
                    :processor => lambda { |val| 
                        classes.find { |cls| val.is_a?( cls ) } && 
                            blk.call( val )
                    }
                )
            end
    
            def self.bg_abstract( meth )
 
                define_method( meth ) do
                    raise "Abstract method #{self.class}##{meth} " \
                          "not implemented"
                end
            end
        
            def self.as_instance( val )
         
                return val if val.is_a?( self )

                BitGirderMethods.not_nil( val, :val )
        
                cd = self.get_class_def
        
                cd.instance_mappers.each do |im|
                    ( res = im.processor.call( val ) ) && ( return res )
                end
        
                raise TypeError, 
                      "Don't know how to convert #{val.class} to #{self}"
            end
        
            def self.install_hash

                cd = self.get_class_def
                define_method( :hash ) { cd.hash_instance( self ) }
            end
        end
    end
end

class BitGirderClass

    include BitGirderMethods
    extend BitGirderMethods

    include BitGirderStructure
end

class BitGirderError < StandardError
    
    include BitGirderMethods
    include BitGirderStructure
end

class BitGirderCliApplication < BitGirderClass

    require 'optparse'

    bg_attr :app_class,
            :description => "Class object to be run as application"

    bg_attr :main_method,
            :description => "Entry method to application",
            :default => :run,
            :processor => lambda { |str| BitGirderMethods::ext_to_sym( str ) }

    def initialize( opts = {} )

        super( opts )
        raise "No such method: #@main_method" unless respond_to?( @main_method )

        @verbose = BitGirderLogger.is_debug_env_set?
    end

    private
    def create_base_parser
        
        p = OptionParser.new

        p.on( "-v", "--[no-]verbose", "Run verbosely" ) { |flag|

            @verbose = flag

            BitGirderLogger.get_logger.level = to_bool( flag ) ? 
                BitGirderLogger::CODE : BitGirderLogger::DEFAULT
        }

        p.on( "-h", "--help", "Print this help" ) { 
            puts p
            exit 0
        }
    end

    private
    def default_to_s( defl )
        
        case defl
            when STDOUT then "stdout"
            when STDIN then "stdin"
            when STDERR then "stderr"
            else defl.to_s
        end
    end

    private
    def get_opt_parser_argv( attr )

        argv = []
        
        switch_str = attr.identifier.to_s.gsub( /_/, "-" )

        if attr.processor == BitGirderAttribute::PROCESSOR_BOOLEAN
            argv << "--[no-]#{switch_str}"
        else
            argv << "--#{switch_str} VAL"
        end

        if desc = attr.description

            # attr.default could be the boolean false, which we still want to
            # display
            unless ( defl = attr.default ) == nil
                desc += " (Default: #{default_to_s( defl )})"
            end

            argv << desc
        end

        argv
    end

    private
    def parse_arg( argh )
 
        attr = argh[ :attr ]
        ident = attr.identifier
        
        prev = argh[ :argh ][ ident ] || attr.default # could be nil either way
        
        val = argh[ :arg ]
#        val = attr.processor.call( val ) if attr.processor

        if attr.is_list
            if prev == nil
                val = [ val ]
            else
                val = ( prev << val )
            end
        end

        argh[ :argh ][ ident ] = val
    end

    private
    def configure_parser
        
        p = create_base_parser
        argh = {}

        cd = BitGirderClassDefinition.for_class( @app_class )
        ( cd.attrs.keys - [ :main_method, :app_class ] ).each do |ident|
            
            attr = cd.attrs[ ident ]
            argv = get_opt_parser_argv( attr )

            p.on( *argv ) { |arg| 
                parse_arg( :arg => arg, :argh => argh, :attr => attr ) 
            }
        end

        [ p, argh ]
    end

    private
    def create_app_obj( argh )
        
        BitGirderClassDefinition.validation_listener =
            lambda { |attr, e| 

                warn( e, "Validation of attr #{attr} failed" ) if @verbose
                cli_opt = sym_to_cli_switch( attr.identifier )
                raise "#{cli_opt}: #{e.message}"
            }

        @app_class.new( argh )
    end

    private
    def make_run_ctx( argv_remain )
        
        { :argv_remain => argv_remain, :verbose => @verbose }
    end

    private
    def fail_run_main( e )
        
        log = BitGirderLogger.get_logger
        
        if @verbose
            log.warn( e, "App failed" )
        else
            STDERR.puts( e.message )
        end

        exit EXIT_FAILURE
    end

    private
    def run_main( app_obj, argv_remain )
        
        meth = app_obj.class.instance_method( @main_method )

        args = 
            case arity = meth.arity
                when 0 then []
                when 1 then [ make_run_ctx( argv_remain ) ]
                else raise "Invalid arity for #{meth}: #{arity}"
            end

        begin
            meth.bind( app_obj ).call( *args )
        rescue Exception => e
            fail_run_main( e )
        end
    end

    public
    def run( argv = ARGV )
 
        p, argh = configure_parser

        begin
            p.parse!( argv = argv.dup )
            app_obj = create_app_obj( argh )
        rescue SystemExit => e
            skip_main = true
        rescue Exception => e

            puts e
            puts p.to_s

            if @verbose || BitGirderLogger.is_debug_env_set?
                puts e.backtrace.join( "\n" ) 
            end

            exit EXIT_FAILURE
        end
 
        run_main( app_obj, argv ) unless skip_main
    end

    def self.run( cls )

        BitGirderMethods::not_nil( cls, :cls )
        self.new( :app_class => cls ).run( ARGV )
    end
end

class ObjectPath
    
    include BitGirderMethods
    extend BitGirderMethods

    attr_reader :parent

    # Not to be called by public methods; all paths should be derived from a
    # parent path or initialized with get_root or get_root_list
    private_class_method :new
    def initialize( parent )
        @parent = parent # parent nil ==> this is a root path
    end

    def descend( elt )
        DictionaryPath.send( :new, self, not_nil( elt, "elt" ) )
    end

    def start_list
        ListPath.send( :new, self, 0 )
    end

    # Returns a path starting with the specified root element
    def self.get_root( root_elt )

        not_nil( root_elt, :root_elt )
        DictionaryPath.send( :new, nil, root_elt )
    end

    def self.get_root_list
        ListPath.send( :new, nil, 0 )
    end

    private
    def collect_path

        res = [ node = self ]

        # This loop is correct on the assumption that the caller obtained this
        # instance by the accepted means
        while node = node.parent 
            res.unshift( node )
        end

        res
    end

    public
    def format( formatter = DefaultObjectPathFormatter.new )
        
        not_nil( formatter, "formatter" )

        formatter.format_path_start( str = "" )
        
        collect_path.each_with_index do |elt, idx|
            
            case elt

                when DictionaryPath 
                    formatter.format_separator( str ) unless idx == 0
                    formatter.format_key( str, elt.key )

                when ListPath then formatter.format_list_index( str, elt.index )

                else raise "Unexpected path type: #{elt.class}"
            end
        end

        str
    end
end

class DictionaryPath < ObjectPath

    attr_reader :key

    def initialize( parent, key )

        super( parent )
        @key = not_nil( key, "key" )
    end
end

class ListPath < ObjectPath
    
    attr_reader :index

    def initialize( parent, index )
        
        super( parent )
        @index = nonnegative( index, "index" )
    end

    def next
        ListPath.send( :new, self.parent, @index + 1 )
    end
end

# Subclasses can override methods to customize as desired
class DefaultObjectPathFormatter
    
    public; def format_path_start( str ); end

    public
    def format_separator( str )
        str << "."
    end

    public
    def format_key( str, key )
        str << key.to_s
    end

    public
    def format_list_index( str, indx )
        str << "[ #{indx} ]"
    end
end

end
end
