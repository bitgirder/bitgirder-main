require 'bitgirder/testing'

require 'bitgirder/core'

require 'time'
require 'thread'

module BitGirder
module Testing

include BitGirder::Core

class TestPhase < BitGirderClass

    BEFORE = :before
    TEST = :test
    AFTER = :after

    RUN_ORDER = [ BEFORE, TEST, AFTER ]

    def self.compare( p1, p2 )
        
        RUN_ORDER.index( p1 ).<=>( RUN_ORDER.index( p2 ) )
    end
end

class Events < BitGirderClass
    
    INVOCATION_STARTING = :invocation_starting

    INVOCATION_COMPLETING = :invocation_completing
end

class TestClassContext < BitGirderClass

    bg_attr :test_class

    public
    def test_methods_for_phase( phase )

        @test_class.instance_methods.select do |meth|
            meth.to_s.start_with?( "#{phase}_" )
        end
    end
end

# public frontend to an AbstractInvocation; passed to test instances as
# appropriate
class TestInvocationCallContext < BitGirderClass

    bg_attr :invocation

    public
    def object
        @invocation.object
    end

    public
    def complete( &blk )
        
        unless blk
            @invocation.send( :impl_complete )
            return
        end

        begin
            blk.call()
            @invocation.send( :impl_complete )
        rescue Exception => e
            @invocation.send( :fail_invocation, e )
        end
    end

    public
    def fail_test( err )
        @invocation.send( :fail_invocation, err )
    end
end
            
class AbstractInvocation < BitGirderClass

    bg_attr :engine
    
    bg_attr :context

    bg_attr :object

    bg_attr :phase

    bg_abstract :impl_name

    # takes TestInvocationCallContext as arg
    bg_abstract :impl_start 

    attr_reader :start_time, :end_time
    
    attr_accessor :on_completion, :error

    private
    def run_queue_add( &blk )
        @engine.run_queue << blk
    end

    public
    def name
        "#{ @context.test_class.name }/#{impl_name}"
    end

    # yields to &comp_blk to set final properties; all code other than time
    # creation is run on run queue thread
    private
    def impl_complete( &comp_blk )

        end_time = Time.now # create now before hitting the run queue

        run_queue_add do

            if @on_completion
 
                @end_time = end_time

                @engine.send_event( Events::INVOCATION_COMPLETING, self )
                comp_blk.call if comp_blk
                
                blk, @on_completion = @on_completion, nil
                @engine.run_queue << blk
            else
                code( "ignoring duplicate completion in #{name}" )
            end
        end
    end

    private
    def fail_invocation( err )
        impl_complete { @error = err }
    end

    # calls f, which should respond to :arity and :call, with the correct number
    # of arguments
    private
    def impl_call_test( f, ctx )

        argv = case ( arity = f.arity )
        when 0 then []
        when 1 then [ ctx ]
        else raise "invalid arity #{arity} in #{f}"
        end

        f.call( *argv )
        impl_complete if argv.empty?
    end

    public
    def start
        
        raise "No object set" unless @object
        
        @start_time = Time.now
        
        run_queue_add do
        
            @engine.send_event( Events::INVOCATION_STARTING, self )

            begin
                ctx = TestInvocationCallContext.new( :invocation => self )
                impl_start( ctx )
            rescue Exception => e
                fail_invocation( e )
            end
        end
    end

    public
    def to_s
        "[#{phase}] #{name}"
    end

    public
    def <=>( inv )
        
        return nil unless inv.is_a?( AbstractInvocation )

        phase_ord = TestPhase.compare( @phase, inv.phase )
        return phase_ord if phase_ord != 0
        
        self.name.<=>( inv.name )
    end
end

class MethodInvocation < AbstractInvocation

    bg_attr :method

    private
    def impl_name
        @method.to_s
    end

    private
    def impl_start( ctx )
        impl_call_test( @object.method( @method ), ctx )
    end
end

class AbstractNamedInvocation < AbstractInvocation

    bg_attr :name_val

    private
    def impl_name
        @name_val
    end
end

class ProcInvocation < AbstractNamedInvocation

    bg_attr :block

    private
    def impl_start( ctx )
        impl_call_test( @block, ctx )
    end
end

class CallObjectInvocation < AbstractNamedInvocation
    
    bg_attr :call_object

    def self.is_call_object( val )
        val.respond_to?( :start_test )
    end

    private
    def impl_start( ctx )
        
        direct = true
        
        if @call_object.respond_to?( :test_context= )
            @call_object.test_context = ctx
            direct = false
        end

        @call_object.start_test
        impl_complete if direct
    end
end

class InvocationSet < BitGirderClass
 
    bg_attr :object

    bg_attr :context

    bg_attr :invocations

    attr_accessor :active
end

# Indicates that an invocation was cancelled, before or while it ran
class TestInvocationCancellationError < BitGirderError

    bg_attr :reason,
            :required => false
end

class UnitTestEngine < BitGirderClass

    bg_attr :filter,
            :required => false
    
    bg_attr :event_handler,
            :required => false

    attr_reader :run_queue

    public
    def send_event( ev, arg )
        
        return unless @event_handler
        @event_handler.call( ev, arg )
    end

    private
    def load_tests

        @test_class_contexts = TestClassMixin.mixed_in_by.map do |cls|
            TestClassContext.new( :test_class => cls )
        end
    end

    private
    def create_method_invocations_for_phase( ctx, obj, phase )
        
        ctx.test_methods_for_phase( phase ).map do |meth|
            
            inv = MethodInvocation.new(
                :engine => self,
                :context => ctx,
                :object => obj,
                :phase => phase,
                :method => meth
            )
        end
    end

    private
    def validate_invocations_for_phase( invs )
        
        invs.inject( {} ) do |h, inv|
            
            if prev = h[ inv.name ]
                raise "duplicate invocations: #{inv}"
            else
                h[ inv.name ] = inv
            end

            h
        end
    end

    private
    def create_invocation_for_factory_result( nm, val, ctx, obj )
        
        opts = {
            :name_val => nm,
            :phase => TestPhase::TEST,
            :engine => self,
            :context => ctx,
            :object => obj
        }

        if val.is_a?( Proc )
            ProcInvocation.new( opts.merge( :block => val ) )
        elsif CallObjectInvocation.is_call_object( val )
            CallObjectInvocation.new( opts.merge( :call_object => val ) )
        else 
            raise "unhandled invocation factory value: #{val} (#{val.class})"
        end
    end

    private
    def apply_invocation_factory( fact, ctx, obj )
        
        invs_by_name = obj.method( fact ).call
        
        res = []

        invs_by_name.inject( [] ) do |arr, pair|
            nm, val = *pair
            arr << create_invocation_for_factory_result( nm, val, ctx, obj )
        end
    end

    private
    def apply_invocation_factories( ctx, obj )
        
        facts = obj.class.instance_methods.select do |nm|
            nm.to_s.start_with?( "invocation_factory_" )
        end
        
        facts.inject( [] ) do |arr, fact|
            arr += apply_invocation_factory( fact, ctx, obj )
        end
    end

    private
    def create_invocations_for_phase( ctx, obj, phase )

        res = []
        
        res += create_method_invocations_for_phase( ctx, obj, phase )

        if phase == TestPhase::TEST
            res += apply_invocation_factories( ctx, obj )
        end

        validate_invocations_for_phase( res )

        res
    end

    private
    def apply_invocation_filter( inv_h )
        inv_h[ TestPhase::TEST ].select! { |inv| @filter.call( inv ) }
    end

    private
    def create_invocations_for_context( ctx, obj )

        res = TestPhase::RUN_ORDER.inject( {} ) do |h, phase|
            
            h[ phase ] = create_invocations_for_phase( ctx, obj, phase )
            h
        end

        apply_invocation_filter( res ) if @filter

        res[ TestPhase::TEST ].empty? ? nil : res
    end

    private
    def create_invocation_set_for_context( ctx )

        obj = ctx.test_class.new
        
        unless invs = create_invocations_for_context( ctx, obj )
            return nil
        end

        InvocationSet.new(
            :context => ctx,
            :object => obj,
            :invocations => invs
        )
    end 

    private
    def create_invocation_sets
        
        @invocation_sets = @test_class_contexts.inject( [] ) do |arr, ctx|
            
            if inv_set = create_invocation_set_for_context( ctx )
                arr << inv_set
            end

            arr
        end
    end

    private
    def run_queue_run
        
        while ( blk = @run_queue.deq ) != self
            blk.call()
        end
    end

    private
    def run_queue_init  
        
        @run_queue = Queue.new
        @queue_runner = Thread.start { run_queue_run }
    end

    private
    def invocation_waits_for_phase( phase )
 
        @invocation_sets.inject( [] ) do |arr, inv_set|

            if inv_set.active
                arr += inv_set.invocations[ phase ].map do |inv|
                    { :invocation => inv, :invocation_set => inv_set }
                end
            end

            arr
        end
    end

    private
    def cancel_invocation_set_with_before_inv_error( inv_set, inv )
        
        ce = TestInvocationCancellationError.new( 
            :reason => inv.error,
            :message => "cancellation due to error in #{inv}: #{inv.error}"
        )

        ( TestPhase::RUN_ORDER - [ TestPhase::BEFORE ] ).each do |phase|
            inv_set.invocations[ phase ].each { |inv| inv.error = ce }
        end
    end

    private
    def complete_phase_invocation( w, waits )
        
        waits.delete( w )

        inv = w[ :invocation ]

        if inv.phase == TestPhase::BEFORE && inv.error

            cancel_invocation_set_with_before_inv_error( 
                w[ :invocation_set ], inv )
        end

        start_next_phase if waits.empty?
    end

    private
    def start_next_phase
        
        if @phases.empty?
            @run_queue << self
            return
        end

        phase = @phases.shift
        waits = invocation_waits_for_phase( phase ).dup

        return start_next_phase if waits.empty?

        waits.each do |w| 
            inv = w[ :invocation ]
            inv.on_completion = lambda { complete_phase_invocation( w, waits ) }
            inv.start
        end
    end

    private
    def start_run

        run_queue_init
        @phases = TestPhase::RUN_ORDER.dup
        @invocation_sets.each { |inv_set| inv_set.active = true }

        start_next_phase
    end

    # ruby (falsely) detects a deadlock if we use join() without a limit, so we
    # instead set a big one and spin until runner exits
    private
    def await_queue_runner
        while ! @queue_runner.join( 100000 ) do; end
    end

    public
    def run

        load_tests
        create_invocation_sets

        start_run
        await_queue_runner
    end        

    public
    def results
        @invocation_sets
    end
end

end
end
