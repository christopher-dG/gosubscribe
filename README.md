# gosubscribe

[![Discord](https://img.shields.io/badge/Discord-invite-blue.svg)](https://discordapp.com/oauth2/authorize?client_id=305550679538401280&scope=bot&permissions=3072)
[![Discord](https://img.shields.io/badge/Discord-o!subscribe-blue.svg)](https://discord.gg/qaUhTKJ)

**Subscribe to osu! users to be notified about their new/updated maps.**

To get started, join the o!subscribe Discord server or invite the bot to your
own with the buttons above.



## Command Reference

If you're brand new, you may want to skip ahead to
[authentication](#authentication) for now.

| Command | Argument(s) | Description | Example |
| :-: | :-: | :-: | :-: |
| `.sub` | `mapper, mapper2, ...` | Subscribe to given mappers. | `.sub pishifat, monstrata` |
| `.unsub` | `username1, username2, ...` | Unsubscribe from given mappers. | `.unsub pishifat, monstrata` |
| `.list` |  | Display your current subscriptions. | `.list` |
| `.count` | `mapper1, mapper2, ...` | Display subscriber counts for given mappers. | `.count pishifat, monstrata` |
| `.top` | `[n=5]` | Display subscriber counts for the top `n` mappers. | `.top 10` |
| `.purge` | | Unsubscribe from all mappers. | `.purge` |
| `.help` | | Link to this reference. | `.help` |

### Authentication

Before you can start using the commands above, you need to complete a quick
registration. If you're a brand new user, you'll generally only need `.init`.


| Command | Argument(s) | Description | Example |
| :-: | :-: | :-: | :-: |
| `.init` | | Initialize as a new user. You should not use this command more than once, even if you're on a different platform than the one you first initialized on. | `.init` |
| `.secret` | | Get your unique secret, required for registering on other platforms. | `.secret` |
| `.register` | `secret` | Register on a new platform, provided that you've initialized  elsewhere. | `.register 1234567890` |

**Note**: On Discord, authentication is done via PM with the bot; these
commands won't work in public channels.

***

**osu! in-game chatbot coming soon™**

*gosubscribe is partially powered by [osu!search](https://osusearch.com).*

*gosubscribe is in no way affiliated with [osu!](https://osu.ppy.sh/home).*
