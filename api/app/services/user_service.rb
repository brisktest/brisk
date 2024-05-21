class UserService
  def self.login(nonce)
    CliLoginAttempt.create_from_nonce(nonce).token
  end
end
