module ApiHelper
  def api_current_user
    User.find(request.active_call.metadata[:current_user_id]) if request.active_call.metadata[:current_user_id]
  end
end
