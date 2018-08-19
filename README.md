# mattermost-plugin-giphy
This Mattermost plugin adds Giphy Integration by creating a `/gif` slash command (no webhooks or additional installation required).

## Requirements
- for Mattermost 5.2 or higher: use v0.2.0 release
- for Mattermost 4.6 to 5.1: use v0.1.x release
- for Mattermost below: unsupported versions (plugins can't create slash commands)

## Installation and configuration
1. Go to the [Releases page](https://github.com/moussetc/mattermost-plugin-giphy/releases) and download the package for your OS and architecture.
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the `.tar.gz` package
3. Go to the `System Console > Plugins > Giphy` configuration page that appeared, and configure the Giphy API key. The default key is the [public beta key, which is 'subject to rate limit constraints'](https://developers.giphy.com/docs/).
4. You can also configure the following settings :
    - rating
    - language (see [Giphy Language support](https://developers.giphy.com/docs/#rendition-guide))
    - display size (see [Giphy rendition guide](https://developers.giphy.com/docs/#rendition-guide))
4. **Activate the plugin** in the `System Console > Plugins Management > Management` page

## Manual configuration
If you need to enable & configure this plugin directly in the Mattermost configuration file `config.json`, for example if you are doing a [High Availability setup](https://docs.mattermost.com/deployment/cluster.html), you can use the following lines (remember to set the API key!):
```json
 "PluginSettings": {
        // [...]
        "Plugins": {
            "com.github.moussetc.mattermost.plugin.giphy": {
                "apikey": "YOUR_GIPHY_API_KEY_HERE", 
                "language": "en",
                "rating": "",
                "rendition": "fixed_height_small"
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

## Usage
The `/gif cute doggo` command will make a post with a GIF from Giphy that matches the 'cute doggo' query. The GIF will be directly posted, using the user's avatar and name.

## Development
Run make vendor to install dependencies, then develop like any other Go project, using go test, go build, etc.

If you want to create a fully bundled plugin that will run on a local server, you can use make `mattermost-jira-plugin.tar.gz`.

## What's next?
- Adding a preview mode to mimick the Slack Giphy integration
- Better testing
