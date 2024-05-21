class ProfileController < ApplicationController
  before_action :authenticate_user!

  def edit; end

  def update
    if current_user.update(user_params)
      redirect_to edit_profile_path, notice: 'Profile updated successfully'
    else
      render :edit
    end
  end

  def user_params
    params.require(:user).permit(:name, :profile_image)
  end
end
