class Mapper

  attr_accessor :username
  attr_accessor :id
  attr_accessor :error

  # Supply at at least one of username or id.
  def initialize(username: '', id: '')
    begin
      if !id.to_s.empty? && !username.empty?
        @id = id.to_s
        @username = username
      elsif !id.to_s.empty?
        @id = id.to_s
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{@id}&type=id"
        response = HTTParty.get(url).parsed_response[0]
        response.key?('username') || raise
        @username = HTTParty.get(url).parsed_response[0]['username']
      elsif !username.empty?
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{username}&type=string"
        response = HTTParty.get(url).parsed_response[0]
        (response.key?('user_id') && response.key?('username')) || raise
        @id = response['user_id']
        @username = response['username']
      else
        raise
      end
    rescue
      @error = true
      return
    end

    begin
      update_db
    rescue
      @error = true
    end

    @error = false
  end

  # Insert a mapper into the database.
  def update_db
    result = DB.exec("SELECT * FROM mappers WHERE mapper_id = #{@id}")
    if result.ntuples == 0
      # Add the mapper to the database.
      DB.exec_prepared('insert_mapper', [@id, @username])
      mapset_ids = request_beatmaps
      if !mapset_ids.empty?
        # Add the mapper's maps to the database.
        cmd = 'INSERT INTO maps(mapper_id, mapset_id) VALUES '
        cmd += mapset_ids.map {|m| "(#{@id}, #{m})"}.join(', ')
        DB.exec(cmd)
      end
    # Todo: Concurrent inserts of the same mapper may be causing duplicate key errors.
    else
      mapper = result.to_a[0]
      if mapper['mapper_name'] != @username
        # Name change: rename the old mapper to the new.
        DB.exec(
          "UPDATE mappers SET mapper_name = '#{@username}' WHERE mapper_id = #{@id}"
        )
      end
    end
  end

  # Get a list of all the mapper's beatmaps.
  def request_beatmaps
    url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&u=#{@id}&type=id"
    return Set.new(HTTParty.get(url).parsed_response.map {|b| b['beatmapset_id']})
  end

  # Get any new maps by the mapper.
  def diff_mapper
    existing_mapsets = DB.exec("SELECT mapset_id FROM maps WHERE mapper_id = #{@id}")
    existing_mapsets.map! {|m| m.values[0]}
    new_mapsets = []

    mapset_ids = request_beatmaps

    mapset_ids.each do |mapset_id|
      if !existing_mapsets.include?(mapset_id)
        new_mapsets.push(mapset_id)
        DB.exec("INSERT INTO maps(mapper_id, mapset_id) VALUES (#{@id}, #{mapset_id})")
      end
    end

    return new_mapsets
  end

  # Get the number of subs for this mapper.
  def subs
    DB.exec("SELECT COUNT(*) FROM subscriptions WHERE mapper_id = #{@id}").values[0][0]
  end
end
