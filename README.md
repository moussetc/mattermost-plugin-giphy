# Mattermost GIF commands plugin (ex 'GIPHY plugin')
This Mattermost plugin adds slash commands to get GIFs from either GIPHY or Gfycat:
- `/gif <keywords>` will post one GIF matching the keywords 
- `/gifs <keywords>` will post a private preview of a GIF matching the keywords, and allows you to shuffle it a number of times before making it public. 
**This command will not work on mobile app until this [Mattermost issue](https://github.com/mattermost/mattermost-mobile/issues/2807) is resolved.**

*No webhooks or additional installation required.*

## COMPATIBILITY
- for Mattermost 5.12 or higher: use v1.1.x release (breaking plugin API change)
- for Mattermost 5.10 to 5.11: use v1.0.x release (possibility to put buttons on ephemeral posts)
- for Mattermost 5.2 to 5.9: use v0.2.0 release
- for Mattermost 4.6 to 5.1: use v0.1.x release
- for Mattermost below: unsupported versions (plugins can't create slash commands)

## Installation and configuration
1. Go to the [Releases page](https://github.com/moussetc/mattermost-plugin-giphy/releases) and download the `.tar.gz` package. Supported platforms are: Linux x64, Windows x64, Darwin x64, FreeBSD x64.
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the `.tar.gz` package
3. Go to the `System Console > Plugins > GIF commands` configuration page that appeared, and choose if you want to use GIPHY (API key required, see below) or Gfycat.
4. If you've chosen GIPHY, configure the Giphy API key. The default key is the [public beta key](https://developers.giphy.com/docs/) which is subject to rate limit constraints and thus might not work at any given time: **it must be changed**.
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
### I can't upload or activate the plugin 
- Is your plugin version compatible with your server version? Check the Compatibility section in the README.
- Make sure you have configured the SiteURL setting correctly in the Mattermost administration panel.
- Check the Mattermost logs (`yourURL/admin_console/logs`) for more detail on why the activation failed.

### Error 'Command with a trigger of `/gif` not found'
This happens when the plugin is not activated, see above section.

### Error 'Unable to get GIF URL'
Start by checking the Mattermost logs (`yourURL/admin_console/logs`) for more detail. Usual causes include:
- Using GIPHY as provider and using the public beta Giphy. The log will looks like: `{"level":"error", ... ,"msg":"Unable to get GIF URL", ... ,"method":"POST","err_where":"Giphy Plugin","http_code":400,"err_details":"Error HTTP status 429: 429 Unknown Error"}`. Solution: get your own GIPHY API key as the default one shouldn't be used in production.
- If your Mattermost server is behind a proxy:
  - If the proxy blocks Giphy and Gfycat: there's no solution besides convincing your security department that accessing Giphy is business-critical.
  - If the proxy allows Giphy and Gfycat: configure your Mattermost server to use your [outbound proxy](https://docs.mattermost.com/install/outbound-proxy.html).

### The picture doesn't load
- Your client (web client, desktop client, etc.) might be behind a proxy that blocks GIPHY or Gfycat. Solution: activate the Mattermost [image proxy](https://docs.mattermost.com/administration/image-proxy.html).

### There are no buttons on the shuffle message
- Check your Mattermost version with the compatibility list at the top of this page.

## Development
To build the plugin:
```
make
```
This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:
```
dist/com.example.my-plugin.tar.gz
```

There is a build target to automate deploying and enabling the plugin to your server, but it requires configuration and http to be installed:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```
Alternatively, if you are running your mattermost-server out of a sibling directory by the same name, use the deploy target alone to unpack the files into the right directory. You will need to restart your server and manually enable your plugin.

## What's next?
- Allow customization of the trigger command
- Command to choose between several GIFS at once (alternative to shuffle)
