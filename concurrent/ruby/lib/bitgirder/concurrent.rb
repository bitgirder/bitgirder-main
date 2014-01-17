require 'bitgirder/core'

module BitGirder
module Concurrent

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

end
end
