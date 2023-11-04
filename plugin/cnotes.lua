vim.api.nvim_create_user_command(
	"Journal",
  function(args)
      require('cnotes').startJournaling(args.args)
  end,
	{ desc = "Open up a journal with today or at a specific date", nargs = "*" }
)

vim.api.nvim_create_user_command("JournalForceSync", function()
  require('cnotes').forceResync()
end, { desc = "Force sync to remote server", nargs = 0 })

vim.api.nvim_create_user_command("JournalSync", function()
  require('cnotes').sync()
end, { desc = "Force sync to remote server", nargs = 0 })
