# frozen_string_literal: true

module ApplicationHelper
  def active_class(link_path)
    current_page?(link_path) ? 'active' : ''
  end
end
