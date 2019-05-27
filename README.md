# mattermost-plugin-gifs
This Mattermost plugin adds slash commands to get GIFs from either GIPHY or Gfycat:
- `/gif <keywords>` will post one GIF matching the keywords 
- `/gifs <keywords>` will post a private preview of a GIF matching the keywords, and allows you to shuffle it a number of times before making it public. 
*No webhooks or additional installation required.*

## COMPATIBILITY
- for Mattermost 5.10 or higher: use v1.x.x release (needs to be abe to put button on ephemeral posts)
- for Mattermost 5.2 to 5.9: use v0.2.0 release
- for Mattermost 4.6 to 5.1: use v0.1.x release
- for Mattermost below: unsupported versions (plugins can't create slash commands)

## Installation and configuration
1. Go to the [Releases page](https://github.com/moussetc/mattermost-plugin-giphy/releases) and download the package for your OS and architecture.
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the `.tar.gz` package
3. Go to the `System Console > Plugins > Giphy` configuration page that appeared, and choose if you want to use GIPHY (API key required, see below) or Gfycat.
4. If you've chosen GIPHY, configure the Giphy API key. The default key is the [public beta key](https://developers.giphy.com/docs/) WHICH IS SUBJECT TO RATE LIMIT CONSTRAINTS AND MIGHT NOT WORK AT ANY GIVEN TIME and **must be changed**.
4. You can also configure the following settings :
    - display size (for both GIPHY and Gfycat, see [Giphy rendition guide](https://developers.giphy.com/docs/#rendition-guide))
    - rating (GIPHY only)
    - language (GIPHY only, see [Giphy Language support](https://developers.giphy.com/docs/#rendition-guide))
4. **Activate the plugin** in the `System Console > Plugins Management > Management` page

## Manual configuration
If you need to enable & configure this plugin directly in the Mattermost configuration file `config.json`, for example if you are doing a [High Availability setup](https://docs.mattermost.com/deployment/cluster.html), you can use the following lines (remember to set the API key!):
```json
 "PluginSettings": {
        // [...]
        "Plugins": {
            "com.github.moussetc.mattermost.plugin.giphy": {
		"provider": "giphy",
                "apikey": "YOUR_GIPHY_API_KEY_HERE", 
                "language": "en",
                "rating": "",
                "rendition": "fixed_height_small"
		"renditionGfycat": "100pxGif"
            },
        },
        "PluginStates": {
            // [...]
            "com.github.moussetc.mattermost.plugin.giphy": {
                "Enable": true
            },
        }
    }
```

## TROUBLESHOOTING
- Is your plugin version compatible with your server version? Check the Compatibility section in the README.
- Make sure you have configured the SiteURL setting correctly in the Mattermost administration panel.
- If you get the following error : `{"level":"error", ... ,"msg":"Unable to get GIF URL", ... ,"method":"POST","err_where":"Giphy Plugin","http_code":400,"err_details":"Error HTTP status 429: 429 Unknown Error"}`: the `429` HTTP status code indicate that you have exceeded the allowed requests for your API key. *Make sure your API is valid for your usage.* This typically happens with the default API key, which musn't be used in production.

## Development
Run make vendor to install dependencies, then develop like any other Go project, using go test, go build, etc.

If you want to create a fully bundled plugin that will run on a local server, you can use make `mattermost-plugin-giphy.tar.gz`.

## What's next?
- Allow customization of the trigger command
- Command to choose between several GIFS at once (alternative to shuffle)
