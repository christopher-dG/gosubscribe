class User

  attr_accessor :id  # Long user id.
  attr_accessor :username
  attr_accessor :disc  # 4-digit id.
  attr_accessor :error

  # Data:
  # - String or integer: user discriminator.
  # - Hash: Hash built from a database entry.
  # - Discordrb::User: Discord user.
  def initialize(data)
    if [String, Integer].include?(data.class)
      begin
        ds = DB[:users].where(:user_disc => data).select(:user_disc, :user_id, :user_name)
        @disc, @id, @username = ds.first.values
      rescue
        @error = true
        return
      end
    elsif data.is_a?(Hash)
      @id = data[:user_id]
      @username = data[:user_name]
      @disc = data[:user_disc]
    else
      @id = data.id
      @username = data.username
      @disc = data.discriminator
    end

    ds = DB[:users].where(:user_disc => @disc)
    if  ds.empty?
      DB.call(:insert_user, :disc => @disc, :id => @id, :name => @username)
    else
      if ds.first[:user_name] != @username  # Update with new username.
        DB[:users].where(:user_disc => @disc).update(:user_name => @username)
      end
    end
    @error = false
  end

  def to_s() "#{@username}:#{@disc}:#{@id}:#{@error}" end

  # Convert a User to a Discordrb::User.
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
  def subscribe(mappers)
    mappers.each do |mapper|
      begin
        DB.call(:subscribe, :disc => @disc, :mapper => mapper.id)
      rescue
        # Assume that it's a duplicate key.
      end
    end
  end

  # Unsubscribe the user from some given mappers.
  # mappers: List of mappers to unsubscribe from.
  def unsubscribe(mappers)
    ds = DB[:subscriptions].where(:user_disc => @disc, :mapper_id => mappers.map {|m| m.id})
    ds.delete
  end

  # Return the usernames of all mappers the user is subscribed to.
  def list
    ds = DB[:mappers].natural_join(:subscriptions).where(:user_disc => @disc)
    ds.map {|m| m[:mapper_name]}.sort_by(&:downcase)
  end

  # Unsubscribe from all mappers.
  def purge() DB[:subscriptions].where(:user_disc => @disc).delete end

end
