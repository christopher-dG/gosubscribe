# gosubscribe

[![Discord](https://img.shields.io/badge/Discord-invite-7289da.svg)](https://discordapp.com/oauth2/authorize?client_id=305550679538401280&scope=bot&permissions=3072)
[![Discord](https://img.shields.io/badge/Discord-o!subscribe-7289da.svg)](https://discord.gg/qaUhTKJ)
[![osu!](https://img.shields.io/badge/osu!-Slow%20Twitch-ff80d5.svg)](https://osu.ppy.sh/users/3172543)

**Subscribe to osu! users to be notified about their new/updated maps.**

To get started, join the o!subscribe Discord server, invite the bot to your
own server, or add the bot as a friend on osu! with the buttons above.

## Command Reference

If you're brand new, you may want to skip ahead to
[authentication](#authentication) for now.

| Command | Argument(s) | Description | Example |
| :-: | :-: | :-: | :-: |
| `.sub` | `mapper1, mapper2, ...` | Subscribe to given mappers. | `.sub pishifat, monstrata` |
| `.unsub` | `mapper1, mapper2, ...` | Unsubscribe from given mappers. | `.unsub pishifat, monstrata` |
| `.list` |  | Display your current subscriptions. | `.list` |
| `.purge` | | Unsubscribe from all mappers. | `.purge` |
| `.count` | `mapper1, mapper2, ...` | Display subscriber counts for given mappers. | `.count pishifat, monstrata` |
| `.top` | `[n=5]` | Display subscriber counts for the top `n` mappers. | `.top 10` |
| `.notifyall` | `y/n` | Enable or disable notifications **all** beatmap updates by subscribed mappers (not just new uploads and ranked status updates). | `.notifyall y` |
| `.message` | `discord/osu!` | Set your preference for where you receive messages. | `.message osu!` |
| `.server` | | Link to the o!subscribe Discord server. | `.server` |
| `.invite` | | Link to a `subscription-bot` Discord invite. | `.invite` |
| `.osu` | | Link to the bot's userpage. | `.osu` |
| `.help` | | Link to this reference. | `.help` |

### Authentication

Before you can start using the commands above, you need to complete a quick
registration. If you're a brand new user, you'll generally only need `.init`.

**Note**: On Discord, authentication is done via PM with the bot; these
commands won't work in public channels.

| Command | Argument(s) | Description | Example |
| :-: | :-: | :-: | :-: |
| `.init` | | Initialize as a new user. You should not use this command more than once, even if you're on a different platform than the one you first initialized on. | `.init` |
| `.secret` | | Get your unique secret, required for registering on other platforms. | `.secret` |
| `.register` | `secret` | Register on a new platform, provided that you've initialized  elsewhere. | `.register MySecret` |

***

*gosubscribe is partially powered by [osu!search](https://osusearch.com).*

*gosubscribe is in no way affiliated with [osu!](https://osu.ppy.sh/home).*
