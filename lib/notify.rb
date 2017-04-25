require_relative 'o!subscribe'


# Send messages about a new map to all subscribers of a given mapper.
def notify(map, mapper)
  cmd = "SELECT u.user_disc, u.user_id, u.user_name FROM users u JOIN subscriptions s "
  cmd += "ON u.user_disc = s.user_disc WHERE s.mapper_id = #{mapper.id}"
  subs = DB.exec(cmd).each do |data|
    user = User.new(data).to_discord_user
    user.pm("[experimental] New map by #{mapper.username}: https://osu.ppy.sh/b/#{map['beatmap_id']}")
    puts("Sent message to #{user.username} for #{mapper.username}'s map.")
  end
end

if __FILE__ == $0
  BOT = setup
  JSON.load(HTTParty.get(SEARCH_URL).parsed_response)['beatmaps'].each do |map|
    mapper_name = map['mapper']
    if mapper_name == 'Akali'
      puts mapper_name
      result = DB.exec("SELECT * FROM mappers WHERE mapper_name = '#{mapper_name}'")
      puts result.to_a
      if result.ntuples > 0
        mapper = Mapper.new(username: mapper_name)
        notify(map, mapper)
        DB.exec("INSERT INTO maps(mapper_id, mapset_id) VALUES (#{mapper.id}, #{map['beatmapset']})")
      end
    end
  end
end
