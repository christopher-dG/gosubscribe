require 'discordrb'
require 'httparty'
require 'pg'
require 'set'

require_relative 'consts'
require_relative 'user'
require_relative 'mapper'

# Return true if a string is empty or comprised of all spaces.
def empty?(s) s.empty? || s.each_char.all? {|c| c == ' '} end

# Escape Markdown underlining.
def escape(s) s.gsub('_', '\_') end

# Respond with an error message.
def failure(event) "Sorry #{event.user.mention}, something went wrong." end

# Subscribe or unsubscribe a user to/from a mapper.
# type: :sub for subscription, :unsub for unsubscription.
def edit_subscription(event, type)
  if event.text.split.length == 1
    msg = failure(event)
  else
    user = User.new(event.user)
    string = event.text.split[1..-1].join(' ')
    tokens = Set.new(string.split(', ').reject {|s| empty?(s)}).map {|t| t.strip}
    mappers = tokens.map {|t| Mapper.new(username: t)}.reject {|m| m.error}
    usernames = mappers.map {|m| m.username}
    if !user.error && !mappers.empty?

      if type == :sub
        if mappers.length > SUB_LIMIT  # Too many subs at once.
          msg = "#{event.user.mention}, you can only subscribe to #{SUB_LIMIT} mappers at once."
        else
          subs = DB.exec("SELECT COUNT(*) FROM subscriptions WHERE user_disc = #{user.disc}")
          subs = subs.values[0][0].to_i
          if subs + mappers.length > TOTAL_SUB_LIMIT  # Too many total subs.
            rem = TOTAL_SUB_LIMIT - subs
            s = rem > 1 ? 's' : ''
            msg = "#{event.user.mention}, you can only subscribe "
            msg += "to #{rem} more mapper#{s} (#{mappers.length} given)."
          else
            user.subscribe(mappers)
            usernames.map! {|u| escape(u)}
            msg = "#{event.user.mention} has subscribed to: #{usernames.join(', ')}."
          end
        end

      elsif type == :unsub
        cmd = "SELECT mapper_name from subscriptions JOIN mappers "
        cmd += "ON subscriptions.mapper_id = mappers.mapper_id "
        cmd += "WHERE user_disc = #{user.disc}"
        subs = DB.exec(cmd).values.map {|v| v[0]}
        mappers.reject! {|m| !subs.include?(m.username)}
        usernames.reject! {|u| !subs.include?(u)}
        usernames.map! {|u| escape(u)}
        user.unsubscribe(mappers)
        usernames.map! {|u| escape(u)}
        msg = "#{event.user.mention} has unsubscribed from: #{usernames.join(', ')}."
      end

    else
      msg = failure(event)
    end
  end
  msg
end


# Set up the bot and its commands.
def setup
  bot = Discordrb::Commands::CommandBot.new(
    token: TOKEN,
    client_id: CLIENT_ID,
    prefix: '!',
    channels: [CHANNEL],
    command_doesnt_exist_message: 'That command does not exist.',

  )
  bot.bucket(:cmd, delay: 1.5)  # Rate limiter.

  bot.command(
    [:subscribe, :sub],
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Subscribe to mappers. `!sub[scribe] user1, user2`',
  ) do |event|
    edit_subscription(event, :sub)
  end

  bot.command(
    [:unsubscribe, :unsub],
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Unsubscribe from mappers. `!unsub[scribe] user1, user2`',
  ) do |event|
    edit_subscription(event, :unsub)
  end

  bot.command(
    :purge,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Unsubscribe from all mappers.',
  ) do |event|
    user = User.new(event.user)
    if user.error
      msg = failure(event)
    else
      user.purge
      msg = "#{event.user.mention} is no longer subscribed to any mappers."
    end
    msg
  end

  bot.command(
    :list,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'List your subscriptions.',
  ) do |event|
    user = User.new(event.user)
    if user.error
      msg = failure(event)
    else
      subscriptions = user.list
      if subscriptions.empty?
        msg ="#{event.user.mention} is not subscribed any mappers."
      else
        subscriptions.map! {|u| escape(u)}
        msg = "#{event.user.mention} is subscribed to: #{subscriptions.join(', ')}."
        if msg.length > CHAR_LIMIT
          # Todo: Actually deal with this.
          msg = "Too many mappers to display, showing as many as possible.\n#{msg}"[0..1995] + ' ...'
        end
      end
    end
    msg
  end

  bot.command(
    :count,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: "Get users' subscriber counts. `!count user1, user2`.",
  ) do |event|
    string = event.text.split[1..-1].join(' ')
    tokens = Set.new(string.split(',').reject {|s| empty?(s)}).map {|t| t.strip}
    mappers = tokens.map {|t| Mapper.new(username: t)}.reject {|m| m.error}
    usernames = mappers.map {|m| m.username}

    # Todo: Deal with mappers not in the table.
    if !mappers.reject {|m| m.error}.empty?
      vals = usernames.map {|u| "'#{u}'"}.join(', ')
      cmd = "SELECT mappers.mapper_name, count(*) subs FROM "
      cmd += "mappers JOIN subscriptions ON "
      cmd += "mappers.mapper_id = subscriptions.mapper_id AND "
      cmd += "mapper_name IN (#{vals}) GROUP BY mappers.mapper_name"

      # For now, just about if anything goes wrong.
      begin
        results = DB.exec(cmd).to_a.sort_by {|x| x['mapper_name'].downcase}
        results.to_a.empty? && raise
      rescue
        msg = failure(event)
      else
        # Todo: Nicer formatting.
        msg = ''
        results.each do |r|
          s = r['subs'].to_i > 1 ? 's' : ''
          msg += "#{r['mapper_name']}: #{r['subs']} subscriber#{s}\n"
        end
      end
    else
      msg = failure(event)
    end
    msg
  end

  return bot
end


if __FILE__ == $0
  puts("DB: #{DB_NAME}")
  puts("Channel: #{CHANNEL}")
  BOT = setup
  BOT.run

  # loop do
  #   # Todo: Check each mapper for new maps.
  #   sleep 3600  # Todo: Need to evaluate how much CPU power this work takes.
  # end

end
