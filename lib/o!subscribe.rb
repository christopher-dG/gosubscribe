# coding: utf-8
require 'date'
require 'discordrb'
require 'httparty'
require 'sequel'
require 'set'
require 'yaml'

require_relative 'consts'
require_relative 'user'
require_relative 'mapper'


# Return true if a string is empty or comprised of all whitespace.
def empty?(s) s.empty? || s =~ /\A\s*\z/ end

# Escape Markdown underlining.
def escape(s) s.gsub('_', '\_') end

# Respond with an error message.
def failure(event, msg: 'something went wrong.')
  "Sorry #{event.user.mention}, #{msg}"
end

# Create a fenced-code table of mappers and their sub counts.
# Mappers: Hash of {mapper_name => sub_count}.
def format_counts(mappers)
  return 'No subscriptions to display.' if mappers.empty?
  msg = ''
  width = mappers.keys.max_by(&:length).length + 3
  mappers.sort_by {|k, v| v.to_i}.reverse.each do |mapper, subs|
    s = subs.to_i != 1 ? 's' : ''
    msg += "#{mapper} #{' ' * (width - mapper.length)}#{subs} subscriber#{s}\n"
  end
  return "```#{msg}```"
end

# Subscribe or unsubscribe a user to/from a mapper.
# type: :sub for subscription, :unsub for unsubscription.
def edit_subscription(event, type)
  return failure(event, msg: 'no mappers were given.') if event.text.split.length == 1

  user = User.new(event.user)
  return failure(event) if user.error

  string = event.text.split[1..-1].join(' ')
  tokens = Set.new(string.split(',').reject {|s| empty?(s)}.map {|t| t.strip})

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
    sub_count = DB[:subscriptions].where(:user_disc => user.disc).count

    if sub_count + mappers.length > TOTAL_SUB_LIMIT
      rem = TOTAL_SUB_LIMIT - sub_count
      s = rem != 1 ? 's' : ''
      msg = failure(
        event, msg: "you can only subscribe to #{rem} more mapper#{s} (#{mappers.length} given)."
      )
    else
      user.subscribe(mappers)
      usernames.map! {|u| escape(u)}
      msg = "#{event.user.mention} has subscribed to: #{usernames.join(', ')}."
    end

  else  # :unsub
    ds = DB[:subscriptions].natural_join(:mappers).where(:user_disc => user.disc)
    subs = ds.map {|s| s[:mapper_name]}
    mappers.reject! {|m| !subs.include?(m.username)}
    if !mappers.empty?
      usernames = usernames.reject {|u| !subs.include?(u)}.map {|u| escape(u)}
      user.unsubscribe(mappers)
      msg = "#{event.user.mention} has unsubscribed from: #{usernames.join(', ')}."
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
    prefix: '.',
    channels: [CHANNEL],
    command_doesnt_exist_message: 'That command does not exist.',

  )
  bot.bucket(:cmd, delay: 1.5)  # Rate limiter.

  bot.command(
    :sub,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Subscribe to mappers.',
    usage: '.sub username1, username2, :userid1, :userid2',
  ) do |event|
    edit_subscription(event, :sub)
  end

  bot.command(
    :unsub,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Unsubscribe from mappers.',
    usage: '.unsub username1, username2, :userid1, :userid2',
  ) do |event|
    edit_subscription(event, :unsub)
  end

  bot.command(
    :list,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'List your subscriptions.',
    usage: '.list',
  ) do |event|
    user = User.new(event.user)

    if user.error
      msg = failure(event)
    else
      subscriptions = user.list

      if subscriptions.empty?
        msg = "#{event.user.mention} is not subscribed any mappers."
      else
        subscriptions.map! {|u| escape(u)}
        msg = "#{event.user.mention} is subscribed to: #{subscriptions.join(', ')}."
        if msg.length > CHAR_LIMIT
          # Todo: Split into multiple messages.
          msg = "Too many mappers to display, showing as many as possible.\n#{msg}"[0..1995] + ' ...'
        end
      end
    end

    msg
  end

  bot.command(
    :purge,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Unsubscribe from all mappers.',
    usage: '.purge',
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
    :count,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Show mappers\' subscriber counts.',
    usage: '.count username1, username2',
  ) do |event|
    string = event.text.split[1..-1].join(' ')
    tokens = Set.new(string.split(',').reject {|s| empty?(s)}).map {|t| t.strip}

    if tokens.empty?
      msg = failure(event, msg: 'you need to supply at least one mapper.')
    elsif tokens.length > CMD_LIMIT
      msg = failure(event, msg: "you can only `.count` #{CMD_LIMIT} mappers at once.")
    else
      mappers = tokens.map {|t| Mapper.new(username: t)}.reject {|m| m.error}
      if !mappers.empty?
        usernames = mappers.map {|m| m.username}
        ids = mappers.map {|m| m.id}
        ds = DB[:mappers].where(:mapper_id => ids).natural_join(:subscriptions).group_and_count(:mapper_name)
        if ds.empty?
          sub_counts = usernames.map {|u| [u, 0]}.to_h
        else
          sub_counts = ds.map {|r| [r[:mapper_name], r[:count]]}.to_h
        end
        msg = format_counts(sub_counts)
      else
        msg = failure(event)
      end
    end

    msg
  end

  bot.command(
    :top,
    bucket: :cmd,
    rate_limit_message: 'Wait %time% seconds.',
    description: 'Show the top `n` mappers and their subscriber counts.',
    usage: '.top [n]',
    max_args: 1,
    arg_types: [Integer],
  ) do |event, max|
    max = max.nil? ? DEFAULT_TOP : [max, TOP_MAX].min
    if max <= 0
      msg = 'Supplied number to `.top` must be greater than 0.'
    else
      ds = DB[:subscriptions].natural_join(:mappers).select(:count, :mapper_name)
      result = ds.group_and_count(:mapper_name).order(:count).reverse.all[0...max]
      msg = format_counts(result.map! {|r| [r[:mapper_name], r[:count]]}.to_h)
    end

    msg
  end

  return bot
end


if __FILE__ == $0
  now = DateTime.now
  puts("DB: #{DB_NAME}")
  puts("Channel: #{CHANNEL}")
  BOT = setup
  BOT.run
end
