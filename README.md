# Cnotes (wip)

Journaling workflow


## Installing
### Binary
This plugin requires the go cli tool in the repo. It's a way to upload saved journals to an sFTP(which in my case happens to be my NAS)

To install, I recommend using [grm](https://github.com/jsnjack/grm) as it can install binaries from the releases page.

`grm install Lilja/cnotes.nvim`

Otherwise, you need to download the latest version for yourself.


### Neovim
#### Lazy.nvim
```lua
{
    "Lilja/cnotes.nvim",
    config = function()
        require('cnotes')
    end
}
```
