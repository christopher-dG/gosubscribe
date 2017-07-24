# coding: utf-8
require_relative '../lib/o!subscribe'

STATUS_MAP = {
  -2 => 'graveyard',
  -1 => 'WIP',
  0 => 'pending',
  1 => 'ranked',
  2 => 'approved',
  3 => 'qualified',
  4 => 'loved',
}

exit if __FILE__ != $0

today = Date.today.to_s
puts("DB: #{DB_NAME}")
puts(today)
BOT = setup
notifications = {}  # user_id -> [map_hash]
mapsets = []  # Mapsets we've already seen.

2.times do |i|  # Look through 500 * 2 recent maps.
  result = HTTParty.get("#{SEARCH_URL}&offset=#{i}").parsed_response
  JSON.load(result)['beatmaps'].each do |map|
    next if mapsets.include?(map['beatmapset_id'])
    mapsets.push(map['beatmapset_id'])

    mapper_name = map['mapper']
    status = map['beatmap_status'].to_i

    if !DB[:mappers].where(:mapper_name => mapper_name).empty?
      mapper = Mapper.new(username: mapper_name)
      next if mapper.error

      ds = DB[:maps].where(
        :mapper_id => mapper.id,
        :mapset_id => map['beatmapset_id'],
      )

      map['old_status'] = !ds.empty? ? ds.first[:status] : nil
      if ds.empty? || status != map['old_status']  # Map is new or updated.
        subs = DB[:users].natural_join(:subscriptions).where(:mapper_id => mapper.id)
        # Add the map to the notifications to be sent out.
        subs.each do |sub|
          if notifications.key?(sub[:user_id])
            notifications[sub[:user_id]].push(map)
          else
            notifications[sub[:user_id]] = [map]
          end
        end
      end

      # Update the DB with the new or updated map.
      if ds.empty?
        puts("Adding #{map['artist']} - #{map['title']} by #{map['mapper']} to DB")
        DB[:maps].insert(
          :mapper_id => mapper.id,
          :mapset_id => map['beatmapset_id'],
          :status => map['beatmap_status'],
        )
      elsif status != map['old_status']
        puts("Updating  #{map['artist']} - #{map['title']} by #{map['mapper']}")
        DB[:maps].where(
          :mapper_id => mapper.id,
          :mapset_id => map['beatmapset_id'],
        ).update(:status => status)
      end
    end
  end
end

# Send a message to each user.
notifications.each do |user_id, maps|
  user = User.new(DB[:users].where(:user_id => user_id).first).to_discord_user
  msg = "Map updates for #{today}:\n"
  maps.each do |map|
    url = "https://osu.ppy.sh/s/#{map['beatmapset_id']}"
    map_str = "#{map['artist']} - #{map['title']} by #{map['mapper']} (#{url})"
    if map['old_status'].nil?
      msg += "New: #{map_str}"
    else
      old = STATUS_MAP[map['old_status']]
      cur = STATUS_MAP[map['beatmap_status']]
      msg += "#{old} → #{cur}: #{map_str}"
    end
    msg += "\n"
  end
  begin
    user.pm(msg)
    puts("Sent message to #{user.username}##{user.discriminator}.")
  rescue
    puts("Sending message to #{user.username}##{user.discriminator} failed.")
  end
end
