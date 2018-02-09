# mattermost-plugin-giphy
This Mattermost plugin adds Giphy Integration by creating a `/gif` slash command (no webhooks required).

## Requirements
- Mattermost 4.6 (to allow plugins to create slash commands) 

## Compilation & Packaging
If your Mattermost server is a `linux amd64`, you can download the [release package](https://github.com/moussetc/mattermost-plugin-giphy/releases) and go to the next step. 
1. Have a Go environnement setup
2. Download this repository
3. Edit the target OS and architecture of the Mattermost server in the `Makefile`
4. Use the command `make dist` to generate the package `plugin.tar.gz`

## Setup
1. Use the `System Console > Plugins Management > Management` page to upload the `.tar.gz`
2. Once the plugin is successfully installed, go to the new `System Console > Plugins > Giphy` configuration page, and  configure the Giphy API key. The default key is the [public beta key, which is 'subject to rate limit constraints'](https://developers.giphy.com/docs/).
4. You can also configure the following settings :
    - rating
    - language
    - display size
4. Activate the plugin in the `System Console > Plugins Management > Management` page

## Usage
The `/gif cute doggo` command will make a post with a GIF from Giphy that matches the 'cute doggo' query. The GIF will be posted using the user's avatar and name.

## What's next?
- Adding a preview mode to mimick the Slack Giphy integration
- Automatic package building for linux x64 install for new release
- Add configuration for command trigger ?
- Better testing
