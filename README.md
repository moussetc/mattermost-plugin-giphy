# mattermost-plugin-giphy
This Mattermost plugin creates a slash command to add Giphy features.

## Requirements
- Mattermost 4.6 (to allow plugins to create slash commands) 
- a Go developpement environnement for the plugin compilation

## Installation
1. Download the sources of this repository
2. Edit the target OS and architecture of the Mattermost server in the `Makefile`
3. Use the command `make dist` to generate the package `plugin.tar.gz`
4. Use the `System Console > Plugins Management > Management` to upload the package
5. Activate the plugin in the Management page

## Configuration
Once the plugin is installed on the server, a `Giphy` configuration page is added to the `System Console > Plugins` menu. 
It contains settings for rating, language and display size.

## Usage
The `/gif cute doggo` command will make a post with a corresponding GIF from Giphy, using the user's avatar and name.

## What's next?
- Adding a preview mode to mimick the Slack Giphy integration
- Offer some ready-to-go plugin packages for classic install (Unix 64bits...)
- Better testing
