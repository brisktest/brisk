if defined?(WillPaginate)
  module WillPaginate
    module ActiveRecord
      module RelationMethods
        def per(value = nil) = per_page(value)
        def total_count = count
        def first_page? = self == first
        def last_page? = self == last
      end
    end

    module CollectionMethods
      alias num_pages total_pages
    end
  end
end
