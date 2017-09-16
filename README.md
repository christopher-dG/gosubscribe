# gosubscribe

[![Discord](https://img.shields.io/badge/Discord-invite-blue.svg)](https://discordapp.com/oauth2/authorize?client_id=305550679538401280&scope=bot&permissions=3072)
[![Discord](https://img.shields.io/badge/Discord-o!subscribe-blue.svg)](https://discord.gg/qaUhTKJ)

**Subscribe to osu! users to be notified about their new/updated maps.**

## Authentication

### Secrets

Due to the eventual goal of making this service available on both
Discord and osu! IRC, some measures need to be taken to prove that you own
the accounts on each platform. When you register with the bot for the first
time on any platform, you'll receive a unique secret which is then used to
register from any other platform.

### Discord

Authentication on Discord works by PMing the bot. If you have not
registered on any other platform, then enter `.init` to receive your secret.
Your Discord ID will also be automatically registered. If you have already
received a secret via another platform (i.e. osu! IRC), then intead use
`.register [your_secret]`. You can retrieve your secret at any time with
the `.secret` command.

### osu! IRC (in-game chat)

`TODO`

## Command Reference

`TODO`

***

*gosubscribe is partially powered by [osu!search](https://osusearch.com).*

*gosubscribe is in no way affiliated with [osu!](https://osu.ppy.sh/home).*
