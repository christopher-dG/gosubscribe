class User

  attr_accessor :bot
  attr_accessor :id  # Long user id.
  attr_accessor :username
  attr_accessor :disc  # 4-digit id.
  attr_accessor :error

  def initialize(data)
    @bot = BOT
    if data.is_a?(String)
      db.exec("SELECT * FROM users WHERE user_disc = #{data}") do |result|
        if result.empty?
          @error = true
        else
          @disc, @id, @username = result.to_a[0].values
        end
      end
    elsif data.is_a?(Hash)
      @id = data['user_id']
      @username = data['user_name']
      @disc = data['user_disc']
    else
      begin
        @id = data.id
        @username = data.username
        @disc = data.discriminator
      rescue
        @error = true
        return
      end
    end

    if DB.exec("SELECT * FROM users WHERE user_disc = #{@disc}").ntuples == 0
      DB.exec_prepared('insert_user', [@disc, @id, @username])
    end
    @error = false
  end

  def to_discord_user
    data = {
      'username' => @username,
      'id' => @id,
      'discriminator' => @disc,
      'avatar_id' => '',
      'bot' => false,
    }
    return Discordrb::User.new(data, BOT)
  end

  # Subscribe the user to some given mappers.
  # mappers: List of mappers to subscribe to.
  def subscribe!(mappers)
    mappers.each do |mapper|
      begin
        DB.exec("INSERT INTO subscriptions(user_disc, mapper_id) VALUES (#{@disc}, #{mapper.id})")
      rescue
        # Assume that it's a duplicate key.
      end
    end
  end

  # Unsubscribe the user from some given mappers.
  # mapper: List of mappers to unsubscribe from.
  def unsubscribe!(mappers)
    mapper_ids = mappers.map {|mapper| mapper.id}
    cmd = "DELETE FROM subscriptions WHERE user_disc = #{@disc} AND mapper_id in "
    cmd += "(#{mapper_ids.join(', ')})"
    DB.exec(cmd)
  end

  # Return the usernames of all mappers the user is subscribed to.
  def list
    names = []
    # Todo: This is definitely not the right way to nest SQL statements.
    DB.exec("SELECT mapper_id FROM subscriptions WHERE user_disc = #{@disc}") do |result|
      result.values.map {|v| v[0]}.each do |mapper_id|
        names.push(
          DB.exec("SELECT mapper_name FROM mappers WHERE mapper_id = #{mapper_id}").values[0][0]
        )
      end
    end
    return names
  end

  # Unsubscribe from all mappers.
  # user: User to be unsubscribed.
  def purge() DB.exec("DELETE FROM subscriptions WHERE user_disc = #{@disc}") end

end
