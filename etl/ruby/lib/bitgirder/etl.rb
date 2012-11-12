require 'bitgirder/core'

module BitGirder
module Etl

include BitGirder::Core

PHASE_EXTRACT = :extract
PHASE_TRANSFORM = :transform
PHASE_LOAD = :load

class BlockScanner < BitGirderClass

    bg_attr :records
    bg_attr :block

    public
    def each_with_id
        @records.each_with_index { |rec, idx| yield( rec, @block.ids[ idx ] ) }
    end
end

class RecordBlock < BitGirderClass
    
    bg_attr :ids
    bg_attr :records
    bg_attr :next_read

    private
    def impl_initialize
        
        id_len = @ids.size

        @records.each_pair do |coding, recs|
            unless recs.size == @ids.size
                raise "Block has #{@ids.size} ids but #{@recs.size} records " \
                      "for coding #@coding"
            end
        end
    end

    public
    def size
        @ids.size
    end

    public
    def coding( nm )
        
        recs = 
            ( @records[ nm ] or raise "Block has no records for coding: #{nm}" )

        BlockScanner.new( :records => recs, :block => self )
    end
end

end
end
