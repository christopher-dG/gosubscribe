class Mapper

  attr_accessor :username
  attr_accessor :id
  attr_accessor :error

  # Supply at at least one of username or id.
  def initialize(username: '', id: '')
    begin
      if !id.empty?
        @id = id
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{@id}&type=id"
        @username = HTTParty.get(url).parsed_response[0]['username']
      elsif !username.empty?
        url = "#{OSU_URL}/get_user?k=#{OSU_KEY}&u=#{username}&type=string"
        response = HTTParty.get(url).parsed_response[0]
        @id = response['user_id']
        @username = response['username']
      else
        raise
      end
    rescue
      @error = true
    end

    # Check if the mapper already has maps in the database.
    if DB.exec("SELECT * FROM mappers WHERE mapper_id = #{@id}").ntuples == 0
      DB.exec_prepared('insert_mapper', [@id, @username])
      mapset_ids = self.request_beatmaps
      if !mapset_ids.empty?
        # Add the mapper's maps to the database.
        cmd = 'INSERT INTO maps(mapper_id, mapset_id) VALUES '
        cmd += mapset_ids.map {|m| "(#{@id}, #{m})"}.join(', ')
        DB.exec(cmd)
      end
    end

    @error = false
  end

  # Get a list of all the Mapper's beatmaps.
  def request_beatmaps
    url = "#{OSU_URL}/get_beatmaps?k=#{OSU_KEY}&u=#{@id}&type=id"
    return Set.new(HTTParty.get(url).parsed_response.map {|b| b['beatmapset_id']})
  end

  # Get any new maps by the mapper.
  def diff_mapper
    existing_mapsets = DB.exec("SELECT mapset_id FROM maps WHERE mapper_id = #{@id}")
    existing_mapsets.map! {|m| m.values[0]}
    new_mapsets = []

    # Todo: use the `since` API parameter to speed this up.
    mapset_ids = self.request_beatmaps

    mapset_ids.each do |mapset_id|
      if !existing_mapsets.include?(mapset_id)
        new_mapsets.push(mapset_id)
        DB.exec("INSERT INTO maps(mapper_id, mapset_id) VALUES (#{self.to_s}, #{mapset_id})")
      end
    end

    return new_mapsets
  end

end
