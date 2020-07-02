# GamerDeathBot
Get up and prevent Gamer Death!

Gamer Death Bot is a Twitch IRC Chatbot that reminds streamers to get up and stretch every 3 hours.

Visit GDB's [channel](https://www.twitch.tv/gamerdeathbot) if you'd like to add or remove it.

## Run It Yourself

### Docker

The simplest way to run GDB is if you have docker-compose installed.

You will also need a PostgreSQL server and credential information setup.

The Postgres DB must have a table 'channels' with 2 fields: 'name' and 'id', both strings. These will be the channels that GDB will listen on.

```bash
git clone https://github.com/kelr/gamerdeathbot
```

Modify ```gdb.env``` with your Twitch Client ID, Secret, DB information and IRC OAuth token.

Run:

```bash
docker-compose -f GamerDeathBot/docker-compose.yml up
```

Prebuilt docker images can be found at [Docker Hub](https://hub.docker.com/r/kyrotobi/gamerdeathbot). These will still require the config environment variables be set.

## Contributions

Any and all contributions or bug fixes are appreciated.