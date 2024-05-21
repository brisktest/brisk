class MembershipsController < ApplicationController
  before_action :authenticate_user!

  def request_access
    @org = Org.find_by_name!(params[:org_name])
  end

  def create_request
    @org = Org.find_by_name!(params[:org_name])
    MembershipMailer.send_request_email(@org, current_user).deliver_later

    redirect_to root_path, notice: 'Your request to join the organization has been sent.'
  end

  # potentially someone could spam this and add a user to an org they don't belong to
  # but it's not a big deal and we can deal with it at a later date if it comes up.
  # I guess you could have an elaborate attack where you convice someone they are
  # operating in your org and not their org.

  def add_user
    @org = Org.find_by_name!(params[:org_name])
    @user = User.find_by_id(params[:user_id])

    redirect_to root_path and return unless current_user.can_manage_org?(@org)

    redirect_to org_path(@org), notice: 'User is already a member of this organization.' if @org.is_member?(@user)

    @membership = Membership.new(org: @org, user: @user, role: 'member')

    if @membership.save
      redirect_to org_path(@org), notice: "User #{@user.email} added to organization."
    else
      redirect_to org_path(@org), notice: 'User could not be added to organization.'
    end
  end

  def claim
    @membership = Membership.find_by!(token: params[:token])

    if @membership.accepted?
      redirect_to root_path,
                  notice: "You have already accepted this membership to the #{@membership.org.name} organization."
      return
    end

    if @membership.is_expired?
      redirect_to root_path, notice: 'This invitation has expired.'
      return
    end

    unless @membership.user_id == current_user.id || current_user.email == @membership.invited_email
      redirect_to root_path, notice: 'You are not allowed to accept this invitation - wrong user.'
      return
    end

    if @membership.cancelled_at.present?
      redirect_to root_path, notice: 'This invitation has been cancelled.'
      return
    end
    @membership.accept! current_user
    redirect_to root_path, notice: "You have joined the #{@membership.org.name} organization."
  end

  def cancel
    @membership = Membership.find_by!(token: params[:token])
    @org = @membership.org
    raise "can't manage this membership" unless current_user.can_manage_org?(@membership.org)

    if @membership.cancel!
      respond_to do |format|
        format.turbo_stream do
          flash.now[:notice] = 'Membership cancelled'
          return
        end
        format.html do
          redirect_to root_path,
                      notice: "You have cancelled this membership to the #{@membership.org.name} organization."
          return
        end
      end
    else
      format.turbo_stream do
        flash.now[:error] = @membership.errors.full_messages.join(', ')

        render_turbo_flash
        return
      end
      format.html do
        redirect_to root_path,
                    notice: "You could not cancel this membership to the #{@membership.org.name} organization."
        return
      end
    end
  rescue StandardError => e
    respond_to do |format|
      format.turbo_stream do
        flash.now[:error] = 'Internal Error - cannot cancel membership'

        render_turbo_flash
        raise e if Rails.env.development?

        return
      end
      format.html do
        redirect_to root_path,
                    notice: "You could not cancel this membership to the #{@membership.org.name} organization."
        return
      end
    end
  end

  def create
    unless params[:membership][:invited_email].present? || params[:membership]['user_id'].present?
      raise 'need an email or a current user'
    end
    raise 'need an org' unless params[:membership][:org_id].present?

    @org = Org.find(params[:membership][:org_id])
    raise 'not allowed' unless current_user.can_manage_org?(@org)

    @membership = @org.memberships.create inviter: current_user, invited_email: params[:membership][:invited_email],
                                          user_id: params[:membership][:user_id]
    @membership.inviter = current_user
    if @membership.save
      MembershipMailer.invite(@membership).deliver_later
      respond_to do |format|
        format.html do
          redirect_to new_org_membership(@org), notice: 'Invitation queued for delivery'
          return
        end
        format.turbo_stream do
          flash.now[:notice] = 'Invitation queued for delivery'
          # render_turbo_flash
          return
        end
      end
    else
      flash.now[:error] = @membership.errors.full_messages.join(', ')
      respond_to do |format|
        format.turbo_stream do
          render_turbo_flash

          return
        end
      end
      redirect_to @org, error: 'Unable to invite', status: :unprocessable_entity
    end
  end

  def new
    @org = Org.find(params[:org_id])
    raise 'not allowed' unless current_user.can_manage_org?(@org)

    @membership = @org.memberships.new
  end

  def change_role
    new_role = params[:new_role].to_sym
    @membership = Membership.find_by!(token: params[:token])
    raise 'not allowed' unless current_user.can_manage_org?(@membership.org)

    @membership.change_role(new_role)
    redirect_to @membership.org, notice: 'Role changed'
  end
end
