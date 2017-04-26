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
def failure(event, msg: 'something went wrong.')
  "Sorry #{event.user.mention}, #{msg}."
end

# Subscribe or unsubscribe a user to/from a mapper.
# type: :sub for subscription, :unsub for unsubscription.
def edit_subscription(event, type)
  return failure(event, msg: 'no mappers were given.') if event.text.split.length == 1

  user = User.new(event.user)
  return failure(event) if user.error

  string = event.text.split[1..-1].join(' ')
  tokens = Set.new(string.split(', ').reject {|s| empty?(s)}).map {|t| t.strip}
  return failure(
           event, msg: "you can only supply up to #{CMD_LIMIT} mappers at once."
         ) if tokens.length > CMD_LIMIT

  mappers = tokens.map do |token|
    if token.start_with?(':')
      Mapper.new(id: token[1..-1])
    else
      Mapper.new(username: token)
    end
  end

  mappers.reject! {|m| m.error}
  usernames = mappers.map {|m| m.username}

  return failure(event) if mappers.empty?

  if type == :sub
    subs = DB.exec("SELECT COUNT(*) FROM subscriptions WHERE user_disc = #{user.disc}")
    subs = subs.values[0][0].to_i

    if subs + mappers.length > TOTAL_SUB_LIMIT  # Too many total subs.
      rem = TOTAL_SUB_LIMIT - subs
      s = rem > 1 ? 's' : ''
      msg = failure(
        event, msg: "you can only subscribe to #{rem} more mapper#{s} (#{mappers.length} given)."
      )
    else
      user.subscribe(mappers)
      usernames.map! {|u| escape(u)}
      msg = "#{event.user.mention} has subscribed to: #{usernames.join(', ')}."
    end

  else  # :unsub
    cmd = 'SELECT mapper_name from subscriptions JOIN mappers '
    cmd += 'ON subscriptions.mapper_id = mappers.mapper_id '
    cmd += "WHERE user_disc = #{user.disc}"
    subs = DB.exec(cmd).values.map {|v| v[0]}
    mappers.reject! {|m| !subs.include?(m.username)}
    usernames = usernames.reject {|u| !subs.include?(u)}.map {|u| escape(u)}
    user.unsubscribe(mappers)
    msg = "#{event.user.mention} has unsubscribed from: #{usernames.join(', ')}."
  end
  return msg
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
    :sub,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: SUB_MSG,
  ) do |event|
    edit_subscription(event, :sub)
  end

  bot.command(
    :unsub,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: UNSUB_MSG,
  ) do |event|
    edit_subscription(event, :unsub)
  end

  bot.command(
    :purge,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: PURGE_MSG,
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
    description: LIST_MSG,
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
    description: COUNT_MSG,
  ) do |event|
    string = event.text.split[1..-1].join(' ')
    tokens = Set.new(string.split(',').reject {|s| empty?(s)}).map {|t| t.strip}

    if tokens.length > CMD_LIMIT
      return msg = failure(event, msg: "you can only !count #{CMD_LIMIT} mappers at once.")
    end

    mappers = tokens.map {|t| Mapper.new(username: t)}.reject {|m| m.error}
    usernames = mappers.map {|m| m.username}

    if !mappers.empty?
      msg = ''
      sub_counts = {}

      mappers.each do |mapper|
        cmd = 'SELECT m.mapper_name, COUNT(*) subs FROM mappers m JOIN '
        cmd += 'subscriptions s ON m.mapper_id = s.mapper_id WHERE '
        cmd += "m.mapper_id = #{mapper.id} GROUP BY m.mapper_name"
        result = DB.exec(cmd).to_a

        if !result.empty?
          sub_counts[mapper.username] = result[0]['subs']
        else
          sub_counts[mapper.username] = 0
        end
      end
      sub_counts.sort_by {|p| p[0].downcase }.each do |pair|
        s = pair[1].to_i != 1 ? 's' : ''
        msg += "#{pair[0]}: #{pair[1]} subscriber#{s}\n"
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
end
