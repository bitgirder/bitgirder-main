require 'bitgirder/core'

require 'irb'

module BitGirder
module Irb

    STATE = {}

    def Irb.run( opts )
        
        STATE[ :init_eval ] = opts[ :init_eval ]
        IRB.start( __FILE__ )
    end

end
end

module IRB

    class Irb
        
        BG_EVAL_INPUT_ORIG = instance_method( :eval_input )
 
        def eval_input( *argv )
            
            if ( script = BitGirder::Irb::STATE[ :init_eval ] )
                @context.evaluate( script, 1 )
            end

            BG_EVAL_INPUT_ORIG.bind( self ).call( *argv )
        end
    end

end
