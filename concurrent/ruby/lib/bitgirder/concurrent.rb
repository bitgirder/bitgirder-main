require 'bitgirder/core'

module BitGirder
module Concurrent

# Classes and modules in here which make use of event machine will call this to
# require it. We don't want to put a bare 'require' statement in this file so as
# to allow for parts of BitGirder::Concurrent to be used without event machine
# being present.
def self.require_em
    require 'eventmachine'
end

class Completion
    
    private_class_method :new

    def initialize( res, ex )
        @res, @ex = res, ex
    end

    public
    def ok?
        @ex == nil
    end

    alias ok ok?
    alias is_ok ok?
    alias is_ok? ok?
    
    public
    def get
        ok? ? @res : ( raise @ex )
    end

    private
    def access( ok_expct, meth_name, ret_val )
        
        ok_actual = ok?

        if ( ok_actual == ok_expct )
            ret_val
        else
            raise "Attempt to call #{meth_name} when ok? returns #{ok_actual}"
        end
    end

    public
    def get_result
        access( true, __method__, @res )
    end

    alias result get_result

    public
    def get_exception
        access( false, __method__, @ex )
    end

    alias exception get_exception
    
    public
    def to_s
        self.inspect
    end

    def self.create_success( res = nil )
        new( res, nil )
    end

    def self.create_failure( ex )

        BitGirder::Core::BitGirderMethods.not_nil( ex, :ex )
        new( nil, ex )
    end
end

class Rendezvous < BitGirder::Core::BitGirderClass
    
    class RendezvousStateError < StandardError; end
    class ClosedError < RendezvousStateError; end
    class UnderflowError< RendezvousStateError; end

    attr_reader :remain

    def initialize( &blk )

        super( {} )

        @on_join = blk # could be nil
        @remain = 0
    end

    public
    def fire
        
        raise ClosedError if closed?
        @remain += 1
    end

    private
    def check_complete

        if @remain == 0 && closed?

            if @on_join
                @on_join.call
            end
        end
    end

    public
    def arrive
        
        if @remain == 0 
            raise UnderflowError
        else
            check_complete if @remain -= 1
        end
    end

    public
    def closed?
        @closed
    end

    public
    def close

        if closed?
            raise ClosedError 
        else
            @closed = true
            check_complete
        end
    end

    public
    def open?
        ! closed?
    end

    # Used by self.run() below to build a call sequence
    class Run
        
        def initialize
            @fires = []
        end

        def complete( &blk )
            @on_join = ( blk or raise "Need a block" )
        end

        def fire( &blk )
            @fires << ( blk or raise "Need a block" )
        end
    end

    def self.run
        
        run = Run.new
        yield( run )

        blk = run.instance_variable_get( :@on_join ) 
        raise "Need a complete block" unless blk

        r = Rendezvous.new( &blk )

        run.instance_variable_get( :@fires ).each { |f| r.fire; f.call( r ) }
        r.close
    end
end

class Retry < BitGirder::Core::BitGirderClass
 
    class Builder

        attr_accessor :retries, :seed_secs
        
        def initialize
            
            # Set defaults
            @retries = 3
            @seed_secs = 1.0
            @failed = lambda { |err| raise err }
            @retry_on = [ Exception ]
        end

        [ :action, :async_action, :complete, :failed ].each do |m|
            class_eval <<-CODE
                # Works as a setter during build and an accessor, used
                # internally by the containing Retry instance
                def #{m}( &blk )
                    @#{m} = blk if blk
                    @#{m}
                end
            CODE
        end

        def retry_on( *argv, &blk )

            if blk
                if argv.empty?
                    @retry_on = blk
                else
                    raise "Can't combine block and rescue target list"
                end
            else
                if argv.empty?
                    raise "Need at least one rescue target"
                else
                    @retry_on = Array.new( argv )
                end
            end
        end
    end

    attr_reader :attempt, :start_time

    def initialize( *argv )
 
        super
        BitGirder::Concurrent.require_em

        @bldr = Builder.new
        yield @bldr if block_given?
        @bldr.freeze

        [ :retries, :seed_secs, :retry_on ].each do |m|
            raise ":#{m} not set" unless @bldr.instance_variable_get( :"@#{m}" )
        end

        if @bldr.action || @bldr.async_action
            if @bldr.action && @bldr.async_action
                raise "Both a synchronous and asynchronous action were set"
            end
        else
            raise "Neither :action nor :async_action was set"
        end

        @attempt = 0
    end

    public
    def seed_secs
        @bldr.seed_secs
    end

    private
    def should_retry?( err )
        
        case val = @bldr.instance_variable_get( :@retry_on )

            when Array then val.find { |err_cls| err.is_a?( err_cls ) }
            else val.call( err )
        end
    end

    private
    def invoke_call( val, call, name )
            
        args = 
            case arity = call.arity
                when 0, -1 then []
                when 1 then [ val ]
                when 2 then [ self, val ]
                else raise "Unexpected arity in #{name}: #{arity}"
            end

        call.call( *args )
    end 

    private
    def fail_final( err )
        invoke_call( err, @bldr.failed, "block :failed" )
    end

    private
    def complete_final( res )
        
        if @bldr.complete
            
            begin
                invoke_call( res, @bldr.complete, "block :complete" )
            rescue Exception => e 
                fail_final( e )
            end
        end
    end

    private
    def action_failed( err )
        
        if res = should_retry?( err )
            
            if @attempt == @bldr.retries
                fail_final( err )
            else
                dur = @bldr.seed_secs * ( 1 << ( @attempt - 1 ) )
                EM.add_timer( dur ) { run_attempt }
            end
        else
            fail_final( err )
        end
    end

    # Callback for an async action
    public
    def complete( res = nil )
        
        if block_given?

            if res == nil
                begin
                    complete_final( yield ) # okay for yield to return nil
                rescue Exception => e
                    action_failed( e )
                end
            else
                raise "Block passed to complete with non-nil result: #{res}"
            end
        else
            complete_final( res )
        end
    end

    public
    def fail_attempt( err )
        complete { raise err }
    end

    private
    def run_attempt

        @attempt += 1

        begin
            if @bldr.action
                complete_final( @bldr.action.call( self ) )
            else
                @bldr.async_action.call( self )
            end
        rescue Exception => e
            action_failed( e )
        end
    end

    public
    def run
        
        if @start_time
            raise "run() already called"
        else
            @start_time = Time.now
            run_attempt
        end
    end

    def Retry.run( *argv, &blk )
        Retry.new( *argv, &blk ).run
    end
end

module EmUtils

    extend BitGirder::Core::BitGirderMethods

    @@did_require = false

    def self.ensure_em   
        unless @@did_require
            Concurrent.require_em
            @@did_require = true
        end
    end

    def defer( callback, &blk )
        
        ensure_em

        not_nil( callback, :callback )
        raise "Need an operation block" unless blk
        
        op = 
            lambda do
                begin
                    Completion.create_success( blk.call )
                rescue Exception => e
                    Completion.create_failure( e )
                end
            end
        
        EM.defer( op, callback )
    end
    module_function :defer
end


end
end
