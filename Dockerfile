FROM ruby:2.4.1

ENV GEMS discordrb httparty pg sequel
ENV APP /root/app

RUN gem install $GEMS && \
    mkdir -p $APP

COPY . $APP

CMD ["ruby", "/root/app/lib/bot.rb"]
