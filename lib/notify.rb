require_relative 'o!subscribe'


# Send messages about a new map to all subscribers of a given mapper.
def notify(map, mapper)
  cmd = "SELECT u.user_disc, u.user_id, u.user_name FROM users u JOIN subscriptions s "
  cmd += "ON u.user_disc = s.user_disc WHERE s.mapper_id = #{mapper.id}"
  DB.exec(cmd).each do |data|
    user = User.new(data).to_discord_user
    begin
      status = map['approved'] == '1' ? 'ranked' : 'qualified'
      user.pm("New #{status} map by #{mapper.username}: https://osu.ppy.sh/b/#{map['beatmap_id']}")
    rescue
      puts("Sending to #{user.username} for #{mapper.username}'s map failed.")
    else
      puts("Sent message to #{user.username} for #{mapper.username}'s map.")
    end
  end
end

if __FILE__ == $0
  now = DateTime.now
  puts("#{now.year}-#{now.month}-#{now.day} #{now.hour}:#{now.minute}")
  BOT = setup
  date = now - 1
  date = "#{date.year}-#{date.month}-#{date.day} #{date.hour}:#{date.minute}"
  url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&since=#{date}"
  mappers = []
  HTTParty.get(url).parsed_response.each do |map|
    mapper_name = map['creator']
    mappers.include?(mapper_name) && next
    mappers.push(mapper_name)
    puts("Mapper: #{mapper_name}.")
    result = DB.exec("SELECT * FROM mappers WHERE mapper_name = '#{mapper_name}'")
    if result.ntuples > 0
      mapper = Mapper.new(username: mapper_name)
      puts("Mapper ID: #{mapper.id}.")
      if DB.exec("SELECT * FROM maps WHERE mapper_id = #{mapper.id} AND mapset_id = #{map['beatmapset_id']}").ntuples == 0
        notify(map, mapper)
        puts("Inserting (#{mapper.id}, #{map['beatmapset_id']}).")
        DB.exec("INSERT INTO maps(mapper_id, mapset_id) VALUES (#{mapper.id}, #{map['beatmapset_id']})")
      end
    end
  end
end
