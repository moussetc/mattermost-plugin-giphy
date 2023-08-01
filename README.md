# Mattermost GIF commands plugin (ex-'GIPHY plugin')

[![Build Status](https://github.com/moussetc/mattermost-plugin-giphy/actions/workflows/ci.yml/badge.svg)](https://github.com/moussetc/mattermost-plugin-giphy/actions/workflows/ci.yml/badge.svg)

**Maintainer:** [@moussetc](https://github.com/moussetc)

A Mattermost plugin to post GIFs from **Giphy or Tenor** with slash commands, available on the official Mattermost Plugin Marketplace.

## Usage

### Plugin v2.0.0 & higher

Use the command `/gif "<keywords>" "<custom caption>"` to search for a GIF and shuffle through GIFs until you find one you like. You can also use `/gif <keywords>` if you don't want to add a custom caption.

Example: first choose a GIF with `/gif "waving cat" "Hello!"` and use the Shuffle button to browse others GIFs:

![demo](assets/demo_preview.png)

You can use the Previous button to go back to previous results:
![demo](assets/demo_preview_with_previous.png)

When you found the perfect GIF, use the Send button to post it: 

![demo](assets/demo_post.png).

*If you prefer having both the `/gif` (post GIF without previewing!) AND `/gifs` (preview and choose GIF before posting) as in the previous versions of the plugin, you can disable the 'Force GIF preview before posting' in the plugin configuration.*

## Compatibility
Use the following table to find the correct plugin version for each Mattermost server version:

| Mattermost server | Plugin release | Incompatibility |
| --- | --- | --- |
| 6.5 and higher | v3.0.x and higher | - |
| 6.0 to 6.4 | v2.0.x and higher | breaking plugin API changes |
| 5.20 to 5.39 | v1.2.x and higher | breaking plugin manifest change |
| 5.12 to 5.19 | v1.1.x | breaking plugin API change |
| 5.10 to 5.11 | v1.0.x | buttons on ephemeral posts |
| 5.2 to 5.9 | v0.2.0 | |
| 4.6 to 5.1 | v0.1.x | |
| below | *not supported* |  plugins can't create slash commands |

## Installation and configuration

**In Mattermost 5.16 and later:**
1. In Mattermost, go to **Main Menu > Plugin Marketplace**.
2. Search for the "GIF Commands" plugin, then click **Install** to install it.
3. Once the installation is completed, click **Configure**. This will take you to System Console to configure the plugin.
4. Choose if you want to use GIPHY (default) or Tenor (both of which requires an API key, see below).
5. **Configure the Giphy or Tenor API key** as explained on the configuration page.
6. You can also configure the following settings :
    - display style (non-collapsable embedded image or collapsable full URL preview)
    - rendition style (GIF size, quality, etc.)
    - rating
    - language (not available for Giphy if random is activated)
    - random (true random is only available for Giphy; for Tenor, the random only applies to the current page of results, meaning you'll need to use Shuffle until a new page of results is loaded in order to see new results even in random mode)
7. **Activate the plugin** in the `System Console > Plugins Management > Management` page

If you are running Mattermost 5.15 or earlier, do not have the Plugin Marketplace enabled or want to install a release that was not published to the Marketplace, follow these steps:
1. Go to the [Releases page](https://github.com/moussetc/mattermost-plugin-giphy/releases) and download the `.tar.gz` package. Supported platforms are: Linux x64, Windows x64, Darwin x64, FreeBSD x64.
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the `.tar.gz` package
3. Go to the `System Console > Plugins > GIF commands` and follow the same configuration steps as for the Marketplace install, displayed from Step 4. on the previous §.

### Configuration Notes in HA

If you are running Mattermost v5.11 or earlier in [High Availability mode](https://docs.mattermost.com/deployment/cluster.html), please review the following:

1. To install the plugin, [use these documented steps](https://docs.mattermost.com/administration/plugins.html#plugin-uploads-in-high-availability-mode)
2. Then, modify the config.json [using the standard doc steps](https://docs.mattermost.com/deployment/cluster.html#updating-configuration-changes-while-operating-continuously) to the following (check the [plugin.json](https://github.com/moussetc/mattermost-plugin-giphy/blob/master/plugin.json) file to see the lists of options for language, rating, rendition, etC.).

```json
 "PluginSettings": {
        // [...]
        "Plugins": {
            "com.github.moussetc.mattermost.plugin.giphy": {
                "displaymode": "embedded",
                "provider": "<giphy or tenor>",
                "apikey": "<your API key from Step 4. above, if you've choosen Giphy or Tenor as your GIF provider>", 
                "language": "en",
                "rating": "none",
                "rendition": "fixed_height_small",
                "renditiontenor": "mediumgif",
                "randomsearch": true,
                "disablepostingwithoutpreview": true
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
- Make sure you have configured the `SiteURL` setting correctly in the Mattermost administration panel.
- Check the Mattermost logs (`yourURL/admin_console/logs`) for more detail on why the activation failed.

### Error 'Command with a trigger of `/gif` not found'
This happens when the plugin is not activated, see above section.

### Error 'Unable to get GIF URL'
Start by checking the Mattermost logs (`yourURL/admin_console/logs`) for more detail. Usual causes include:
- Using GIPHY as provider and using the public beta Giphy. The log will looks like: `{"level":"error", ... ,"msg":"Unable to get GIF URL", ... ,"method":"POST","err_where":"Giphy Plugin","http_code":400,"err_details":"Error HTTP status 429: 429 Unknown Error"}`. Solution: get your own GIPHY API key as the default one shouldn't be used in production.
- If your Mattermost server is behind a proxy:
  - If the proxy blocks Giphy and Tenor: there's no solution besides convincing your security department that accessing Giphy is business-critical.
  - If the proxy allows Giphy and Tenor: configure your Mattermost server to use your [outbound proxy](https://docs.mattermost.com/install/outbound-proxy.html).

### The picture doesn't load
- Your client (web client, desktop client, etc.) might be behind a proxy that blocks GIPHY or Tenor. Solution: activate the Mattermost [image proxy](https://docs.mattermost.com/administration/image-proxy.html).
- If the Display Mode configured is "Collapsable Image Preview", then the link previews option must be configured in the System Console (> Posts > Enable Link Previews). Do note that user can also change this option in their Account Settings. 

### There are no buttons on the shuffle message
- Check your Mattermost version with the compatibility list at the top of this page.

## Development

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options. In order for the below options to work, you must first enable plugin uploads via your config.json or API and restart Mattermost.

```json
    "PluginSettings" : {
        ...
        "EnableUploads" : true
    }
```
T### Deploying with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. Edit your server configuration as follows:

```json
{
    "ServiceSettings": {
        ...
        "EnableLocalMode": true,
        "LocalModeSocketLocation": "/var/tmp/mattermost_local.socket"
    },
}
```

and then deploy your plugin:
```
make deploy
```

You may also customize the Unix socket path:
```
export MM_LOCALSOCKETPATH=/var/tmp/alternate_local.socket
make deploy
```

If developing a plugin with a webapp, watch for changes and deploy those automatically:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make watch
```

### Deploying with credentials

Alternatively, you can authenticate with the server's API with credentials:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

## How do I share feedback on this plugin?

Feel free to create a GitHub issue or to contact me at `@cmousset` on the [community Mattermost instance](https://pre-release.mattermost.com/) to discuss.
