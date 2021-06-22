# Obsvraeously
Obsvraeously is a simple discord bot that watches for Avrae (https://avrae.io/) dice rolls and exports them to an OBS layer. This documentation assumes some level of technical skill and are written as if you're using mac OS.

## Create a bot account
Follow these instructions to create a discord bot and add it to your server. The necessary permissions integer is 0. https://discordpy.readthedocs.io/en/stable/discord.html

## Startup the bot
If you have not done so already install docker.


From the command line within the root of the repository.

```
OBSVRAE_DISCORD=<YOUR_BOT_TOKEN> docker compose up -d
```

If you have not previously run the bot this will compile necessary images and start the necessary containers.

If you make code changes you must run

```
docker compose build && OBSVRAE_DISCORD=<YOUR_BOT_TOKEN> docker compose up -d
```

## Add to OBS
In OBS add a new browser source with the following config.

* URL: http://localhost:8080/static/index.html
* Width: 624
* Height: 432