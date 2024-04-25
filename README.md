# cormorant
Source code for the tymas discord server bot, which has /join /leave /color and /weather commands. Join and leave are for pingable roles for coordination of games and the like or updates on particular topics. Color is for assigning nick colors. Weather provides current conditions, today's forecast, and the weekly forecast using open-meteo.

## Running via `docker-compose`
1. Make a `secrets.env` in the following format:
```
appid=yourappid
authtoken=yourtoken
```
2. Copy the included `docker-compose.yml` file from the root of this repo
3. Replace the `env_file` path with the path to your `secrets.env`
4. `docker-compose up`
