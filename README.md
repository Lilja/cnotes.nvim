# Cnotes (wip)

Journaling workflow of locally creating a journal file and uploading it to a remote target.

## Usage
* `:Journal` Enter a journaling mode

Supports varargs such as `3 days ago`, `1 month ago`:

`:Journal yesterday`, `:Journal 4 days ago`

* `:JournalForceSync` If you need to write after upload


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
    -- configuration needed!
    config = function()
        require('cnotes').setup({
            -- The ssh host in your ~/.ssh/config to connect to.
            sshHost = "",
            -- The directory to locally store journals
            localFileDirectory = "~/Documents/journal",
             -- Binary
            syncBinary = "cnotes-sftp-client",
            -- The path on the remote sFTP host of where you want to store files.
            destination = "/",
            -- When using :Journal, what kind of file type it will be.
            fileExtension = ".md",
        })
    end
}
```
