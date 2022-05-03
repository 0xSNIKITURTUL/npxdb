# NPXDB

An ETL tool and Vercel serverless function to extract and load nopixel character and
streamer info from [Hasroot](https://nopixel.hasroot.com/characters.php) and make the
data available to services like twitch bots through a plain text api.

## Fossabot Integration

Create a new command with this value:

`$(customapi https://nopixel-db.vercel.app/api/characters?name=$(query))`

## Usage

### ETL Tool

Download the latest version of NopixelDB (ETL Tool) from the releases section. Make a .env
file in the same directory as the tool. Add a line with `DATABASE_URL="<your postgres url>"`
to the .env file. Run the tool with `./NopixelDB` (\*nix) or `.\NopixelDB` (windows).

### Vercel Function
